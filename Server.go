package fastRPC

import (
	"encoding/json"
	"fastRPC/internal/frpcsync"
	"fmt"
	"github.com/golang/snappy"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type CompressAlgorithm string

const MaxConnectionNumbers = 1024
const MaxTransportByte = 1460

var (
	defaultCompression CompressAlgorithm = "snappy"
)

type fastRPCServer struct {
	server    *net.TCPListener
	RpcServer RPCServer
	Options   *ServerOption
	//grpc is abandoned
	handlerMethod map[string]reflect.Value
	conns         map[string]map[net.Conn]bool
	connsMutex    map[net.Conn]*sync.Mutex
	wg            sync.WaitGroup
	mutex         sync.Mutex
	quit          frpcsync.Event
	done          frpcsync.Event
	drain         bool
	//conns concurrency read write lock
	rwmutex sync.RWMutex
}
type RPCServer interface {
}

var defaultServerOption = &ServerOption{
	MaxConnNums:       MaxConnectionNumbers,
	CompressAlgorithm: defaultCompression,
}

type ServerOption struct {
	MaxConnNums       int
	CompressAlgorithm CompressAlgorithm
	TransportWays     string
}

func NewFastRPCServer(server RPCServer, option *ServerOption) (result *fastRPCServer) {
	result = &fastRPCServer{
		server:  &net.TCPListener{},
		Options: defaultServerOption,
	}
	if option != nil {
		result.Options = option
	}
	result.registerMethod(server)
	return result
}
func (fastRPC *fastRPCServer) Run() error {
	defer func() {
		fastRPC.wg.Wait()
		if fastRPC.quit.HasFired() {
			<-fastRPC.done.Done()
		}
	}()
	defer func() {
		fastRPC.mutex.Lock()
		if fastRPC.server != nil {
			fastRPC.server.Close()
		}
		fastRPC.mutex.Unlock()
	}()
	var timeoutWait time.Duration
	for {
		accept, err := fastRPC.server.Accept()

		if err != nil {
			if _, ok := err.(syscall.Errno); ok {
				if timeoutWait == 0 {
					timeoutWait = 5 * time.Millisecond
				} else {
					timeoutWait *= 2
				}
				if max := 1 * time.Second; timeoutWait > max {
					timeoutWait = max
				}
				timer := time.NewTimer(timeoutWait)
				select {
				case <-timer.C:
				case <-fastRPC.quit.Done():
					timer.Stop()
					return nil
				}
				continue
			}
			if fastRPC.quit.HasFired() {
				return nil
			}
			return err
		}
		fastRPC.wg.Add(1)
		go func() {
			fastRPC.handlerConn(accept)
			fastRPC.wg.Done()
		}()
	}
}

func (fastRPC *fastRPCServer) addConn(addr string, conn net.Conn) bool {
	fastRPC.rwmutex.Lock()
	defer fastRPC.rwmutex.Unlock()
	if fastRPC.conns == nil {
		return false
	}
	if fastRPC.drain {
		return false
	}
	if fastRPC.conns[addr] == nil {
		fastRPC.conns[addr] = make(map[net.Conn]bool)
	}
	//open the read write barrier to defeat instruction rearrangement
	var params int32
	if _, ok := fastRPC.connsMutex[conn]; !ok {
		_ = atomic.LoadInt32(&params)
		fastRPC.connsMutex[conn] = &sync.Mutex{}
		atomic.StoreInt32(&params, 1)
	}
	fastRPC.conns[addr][conn] = true
	return true

}

/**
  support stream data
*/
func (fastRPC *fastRPCServer) handlerConn(conn net.Conn) {
	if fastRPC.quit.HasFired() {
		fastRPC.closeConn(conn.LocalAddr().String(), conn)
		return
	}
	wg := sync.WaitGroup{}
	//forever can't time out
	conn.SetDeadline(time.Time{})
	fastRPC.addConn(conn.LocalAddr().String(), conn)
	defer fastRPC.closeConn(conn.LocalAddr().String(), conn)
	for {
		var data = make([]byte, MaxTransportByte)
		_, err := conn.Read(data)
		if err != nil {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := &Response{
				Code: 300,
			}
			defer func() {
				// recover panic to resolve the error which comes from the rpc executing
				if panicError := recover(); panicError != any(nil) {
					mutex := fastRPC.connsMutex[conn]
					mutex.Lock()
					resp, _ := fastRPC.sendInterceptor(result)
					conn.Write(resp)
					mutex.Unlock()
				}
			}()
			target, err := fastRPC.receiveInterceptor(data)
			//promise call function must exist
			if _, ok := fastRPC.handlerMethod[target.Method]; !ok {
				panic(any(err))
			}
			callFunc, err := fastRPC.callFunc(target.Method, target.Data)
			if err != nil {
				panic(any(err))
			}
			result.GetSuccessResp(callFunc)
			interceptor, err := fastRPC.sendInterceptor(result)
			if err != nil {
				panic(any(err))
			}
			mutex := fastRPC.connsMutex[conn]
			mutex.Lock()
			conn.Write(interceptor)
			mutex.Unlock()
		}()
	}
	wg.Wait()
}
func (fastRPC *fastRPCServer) closeConn(addr string, conn net.Conn) {
	fastRPC.rwmutex.Lock()
	defer fastRPC.rwmutex.Unlock()
	if _, ok := fastRPC.conns[addr]; !ok {
		return
	}
	if _, ok := fastRPC.conns[addr][conn]; !ok {
		return
	}
	conn.Close()
	delete(fastRPC.connsMutex, conn)
	delete(fastRPC.conns[addr], conn)
}

// analysis data to golang struct
func (fastRPC *fastRPCServer) receiveInterceptor(data []byte) (result *DataStandards, err error) {
	var encodingData = make([]byte, MaxTransportByte)
	decode, err := snappy.Decode(encodingData, data)
	if err != nil {
		return nil, fmt.Errorf("error analy the compression data,only support the snappy ")
	}
	if len(decode) > len(encodingData) {
		encodingData = decode
	}
	result = &DataStandards{}
	err = json.Unmarshal(encodingData, result)
	if err != nil {
		return nil, fmt.Errorf("cant parse data by json format")
	}
	return result, nil
}

func (fastRPC *fastRPCServer) registerMethod(server RPCServer) {
	value := reflect.ValueOf(server)
	method := value.NumMethod()
	if fastRPC.handlerMethod == nil {
		fastRPC.handlerMethod = make(map[string]reflect.Value)
	}
	for i := 0; i < method; i++ {
		method := value.Method(i)
		//promise cant repeat
		fastRPC.handlerMethod[method.String()] = method
	}
}
func (fastRPC *fastRPCServer) sendInterceptor(resp *Response) ([]byte, error) {
	marshal, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	var result []byte
	var bytes = make([]byte, MaxTransportByte)
	decode, err := snappy.Decode(bytes, marshal)
	if err != nil {
		return nil, err
	}
	if len(decode) > len(bytes) {
		result = decode
	} else {
		result = bytes
	}
	return result, nil
}
func (fastRPC *fastRPCServer) callFunc(methodName string, params []byte) ([]byte, error) {
	methodParams := strings.Split(string(params), " ")
	methodCall := fastRPC.handlerMethod[methodName]
	convertValue := make([]reflect.Value, len(methodParams))
	for i := 0; i < len(params); i++ {
		convertValue[i] = reflect.ValueOf(methodParams[i])
	}
	funcInputNums := methodCall.Type().NumIn()
	var returnValues []reflect.Value
	if funcInputNums < len(methodParams) {
		returnValues = methodCall.CallSlice(convertValue)
	} else {
		returnValues = methodCall.Call(convertValue)
	}
	funcOutputNums := methodCall.Type().NumOut()
	var result []byte = make([]byte, 0, MaxTransportByte)
	if len(returnValues) > funcOutputNums {
		//exist slice
	} else {
		for i := 0; i < len(returnValues); i++ {
			toByte, err := convertValueToByte(returnValues[i])
			if err != nil {
				return nil, err
			}
			result = append(result, combineToJson(returnValues[i].Type().Name(), string(toByte), returnValues[i].Kind())...)
		}
	}
	return result, nil
}

func convertValueToByte(value reflect.Value) (result []byte, err error) {
	defer func() {
		if panicErr := recover(); panicErr != any(nil) {
			result = nil
			fmt.Println(panicErr)
			err = fmt.Errorf("params analy error")
		}
	}()
	result = make([]byte, 0, 64)
	switch value.Kind() {
	case reflect.String:
		result = append(result, []byte(value.String())...)
	case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32:
		result = append(result, []byte(strconv.FormatInt(value.Int(), 10))...)
	case reflect.Slice:
		var tempArr = make([]byte, 0, value.Len()*value.Type().Elem().Align())
		for i := 0; i < value.Len(); i++ {
			temp, err := convertValueToByte(value.Index(i))
			if err != nil {
				return nil, err
			}
			if i != value.Len()-1 {
				temp = append(temp, ',')
			}
			tempArr = append(tempArr, temp...)
		}
		result = append(result, combineToJson(value.Type().Name(), string(tempArr), value.Kind())...)
	case reflect.Struct:
		var number = value.NumField()
		var temp = make([]byte, 0, value.Type().Size())
		for i := 0; i < number; i++ {
			var fieldValue = value.Field(i)
			bytes, err := convertValueToByte(fieldValue)
			if err != nil {
				return nil, err
			}
			temp = append(temp, combineToJson(fieldValue.Type().Name(), string(bytes), fieldValue.Kind())...)
		}
		result = append(result, combineToJson(value.Type().Name(), string(temp), value.Kind())...)
	case reflect.Bool:
		result = append(result, strconv.FormatBool(value.Bool())...)
	case reflect.Pointer, reflect.Interface, reflect.Uintptr:
		elem := value.Elem()
		temp, err := convertValueToByte(elem)
		if err != nil {
			return nil, err
		}
		result = append(result, temp...)
	case reflect.Float32, reflect.Float64:
		result = append(result, strconv.FormatFloat(value.Float(), 'E', -1, 32)...)
	case reflect.Complex64, reflect.Complex128:
		result = append(result, strconv.FormatComplex(value.Complex(), 'E', -1, 32)...)
	//maybe can improve
	case reflect.Map:
		keys := value.MapKeys()
		if len(keys) <= 0 {
			return
		}
		var tempArr = make([]byte, 0, 2*len(keys)*(keys[0].Type().Align())*value.MapIndex(keys[0]).Type().Align())
		for i := 0; i < len(keys); i++ {
			var keyValue = value.MapIndex(keys[0])
			keyBytes, err := convertValueToByte(keys[i])
			if err != nil {
				return nil, err
			}
			valueBytes, err := convertValueToByte(keyValue)
			if err != nil {
				return nil, err
			}
			tempArr = append(tempArr, combineToJson(string(keyBytes), string(valueBytes), reflect.String)...)
		}
		result = append(result, combineToJson("", string(tempArr), reflect.Struct)...)
	}
	return result, nil
}

func combineToJson(key string, value string, valueType reflect.Kind) []byte {
	sb := strings.Builder{}
	sb.WriteString("\"")
	sb.WriteString(key)
	sb.WriteString("\": ")
	switch valueType {
	case reflect.String:
		sb.WriteString("\"")
		sb.WriteString(value)
		sb.WriteString("\"")
	case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Bool:
		sb.WriteString(value)
	case reflect.Slice, reflect.Array:
		sb.WriteString("[")
		sb.WriteString(value)
		sb.WriteString("]")
	case reflect.Struct:
		sb.WriteString("{")
		sb.WriteString(value)
		sb.WriteString("}")
	}
	sb.WriteString("\n")
	return []byte(sb.String())
}

func (fastRPC *fastRPCServer) Stop() {
	fastRPC.quit.Fire()
	defer func() {
		fastRPC.wg.Wait()
		fastRPC.done.Fire()
	}()
}
