package fastRPC

import (
	"encoding/json"
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
	server, _ := NewFastRPCServer(&customRPC{}, nil)
	err := server.Run()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(server)
}
func TestCallFunc(t *testing.T) {
	server, _ := NewFastRPCServer(&customRPC{}, nil)
	result, err := server.callFunc("HandlerHelloWorld", []byte("hello hello hello"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(result))
}

type test struct {
	Name       string
	Age        int
	Sex        int
	result     map[string]interface{}
	structTest test2
}
type test2 struct {
	Name string
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
		Name:   "xiyang",
		Age:    47,
		Sex:    1,
		result: map[string]interface{}{"xiyang": "hello"},
		structTest: test2{
			Name: "hello",
		},
	}
	value, err := convertValueToByte(reflect.ValueOf(&t2))
	sb := strings.Builder{}
	sb.WriteString("{")
	sb.WriteString(string(value))
	sb.WriteString("}")
	println("is json? ", json.Valid([]byte(sb.String())))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	println(sb.String())
}
