package fastRPC

import (
	"bufio"
	"io/ioutil"
	"net"
	"reflect"
)

const MaxConnectionNumbers = 1024
const CompressAlgorithm = "snappy"

type fastRPCServer struct {
	server        *net.TCPListener
	RpcServer     RPCServer
	Options       *ServerOption
	//grpc is abandoned
	handlerMethod map[string]any
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
		server:&net.TCPListener{},
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
		accept.
	}
}
// analysis data to golang struct
func(fastRPC *fastRPCServer)receiveBefore(){

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
func()sendBefore() {

}
func (fastRPC *fastRPCServer) callFunc(methodName string, dataSource bufio.Reader) {
	all, _ := ioutil.ReadAll(&dataSource)

}
