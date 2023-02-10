package fastRPC

import (
	"fmt"
	"reflect"
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

type test struct {
	Name string
	Age  int
	Sex  int
}

func TestConvertValueSlice(t *testing.T) {
	slice := make([]int, 0, 100)
	for i := 0; i < cap(slice); i++ {
		slice = append(slice, 10)
	}
	value, err := convertValueToByte(reflect.ValueOf(slice))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(value))
}
func TestConvertValueStruct(t *testing.T) {
	t2 := test{
		Name: "xiyang",
		Age:  47,
		Sex:  1,
	}
	value, err := convertValueToByte(reflect.ValueOf(t2))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	println(string(value))
}
