package fastRPC

import (
	"bufio"
	"github.com/golang/snappy"
	"io/ioutil"
	"net"
	"reflect"
	"sync"
)

const MaxConnectionNumbers = 1024
const CompressAlgorithm = "snappy"
const MaxTransportByte = 1460

type fastRPCServer struct {
	server    *net.TCPListener
	RpcServer RPCServer
	Options   *ServerOption
	//grpc is abandoned
	handlerMethod map[string]any
	conns         map[string]map[net.Conn]bool
	wg            sync.WaitGroup
}
type RPCServer interface {
}

var defaultServerOption = &ServerOption{
	MaxConnNums:       MaxConnectionNumbers,
	CompressAlgorithm: CompressAlgorithm,
}

type ServerOption struct {
	MaxConnNums       int
	CompressAlgorithm string
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
func (fastRPC *fastRPCServer) Run() {
	for {
		accept, err := fastRPC.server.Accept()
		if err != nil {
			return
		}
		fastRPC.wg.Add(1)
		go func() {
			fastRPC.handlerConn(accept)
			fastRPC.wg.Done()
		}()
	}
}

func (fastRPC *fastRPCServer) handlerConn(conn net.Conn) {
	var data = make([]byte, MaxTransportByte)
	read, err := conn.Read(data)
	if err != nil {
		return
	}

}

// analysis data to golang struct
func (fastRPC *fastRPCServer) receiveInterceptor(data []byte) (DataStandards, error) {
	var encodingData = make([]byte, MaxTransportByte)
	decode, err := snappy.Decode(encodingData, data)
	if err != nil {
		return
	}
}

func (fastRPC *fastRPCServer) registerMethod(server RPCServer) {
	value := reflect.ValueOf(server)
	method := value.NumMethod()
	for i := 0; i < method; i++ {
		method := value.Method(i)
		//promise cant repeat
		fastRPC.handlerMethod[method.String()] = nil
	}
}
func (fastRPC *fastRPCServer) sendInterceptor(standards *DataStandards) []byte {

}
func (fastRPC *fastRPCServer) callFunc(methodName string, dataSource bufio.Reader) {
	all, _ := ioutil.ReadAll(&dataSource)

}
