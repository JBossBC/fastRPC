package fastRPC

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func BenchmarkJsonConvert(b *testing.B) {
	t2 := test{
		Name: "xiyang",
		Age:  47,
		Sex:  1,
		//result: map[string]interface{}{"xiyang": "hello"},
		//structTest: test2{
		//	Name: "hello",
		//},
	}
	var mapping = make(map[uintptr][]byte)
	for i := 0; i < b.N; i++ {
		value := reflect.ValueOf(t2)
		if _, ok := mapping[uintptr(unsafe.Pointer(&value))]; ok {
			continue
		}
		bytes, err := convertValueToByte(value)
		mapping[uintptr(unsafe.Pointer(&value))] = bytes
		if err != nil {
			fmt.Println("method err ")
			return
		}
		//sb := strings.Builder{}
		//sb.WriteString("{")
		//sb.WriteString(string(value))
		//sb.WriteString("}")
	}
}
func BenchmarkJsonMarshal(b *testing.B) {
	t2 := test{
		Name: "xiyang",
		Age:  47,
		Sex:  1,
		//result: map[string]interface{}{"xiyang": "hello"},
		//structTest: test2{
		//	Name: "hello",
		//},
	}
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(t2)
		if err != nil {
			fmt.Println("system convert error")
			return
		}
	}
}
