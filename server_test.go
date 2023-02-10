package fastRPC

import (
	"fmt"
	"strings"
	"testing"
)

type customRPC struct {
}

func (rpc *customRPC) HandlerHelloWorld(args ...string) string {
	builder := strings.Builder{}
	builder.WriteString("Hello World")
	for i := 0; i < len(args); i++ {
		builder.WriteString(args[i])
	}
	return builder.String()
}

func TestNewFastRPCServer(t *testing.T) {
	server := NewFastRPCServer(&customRPC{}, nil)
	err := server.Run()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(server)
}
func TestConvert