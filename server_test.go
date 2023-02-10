package fastRPC

import (
	"fmt"
	"strings"
	"testing"
)

type customRPC struct {
	RPCServer
}

func (rpc *customRPC) handlerHelloWorld(args ...string) string {
	builder := strings.Builder{}
	builder.WriteString("Hello World")
	for i := 0; i < len(args); i++ {
		builder.WriteString(args[i])
	}
	return builder.String()
}

func TestNewFastRPCServer(t *testing.T) {
	server := NewFastRPCServer(&customRPC{}, nil)
	fmt.Println(server)
}
