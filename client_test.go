package fastRPC

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	dial, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return
	}
	var data = DataStandards{
		Header: nil,
		Method: "HandlerHelloWorld",
		Data:   []byte("hello hello hello"),
	}
	marshal, err := json.Marshal(data)
	if err != nil {
		return
	}
	const times = 1000
	for i := 0; i < times; i++ {
		dial.Write(marshal)
		var result = make([]byte, 1024)
		_, err = dial.Read(result)
		if err != nil {
			return
		}
		ending := findBytesEnding(result)
		var resp Response
		err := json.Unmarshal(result[:ending], &resp)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(string(resp.Result))
		time.Sleep(3 * time.Second)
	}
}
func BenchmarkNewFastRPCServer(b *testing.B) {
	dial, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return
	}
	var data = DataStandards{
		Header: nil,
		Method: "HandlerHelloWorld",
		Data:   []byte("hello hello hello"),
	}
	marshal, err := json.Marshal(data)
	if err != nil {
		return
	}
	for i := 0; i < b.N; i++ {
		dial.Write(marshal)
		var result = make([]byte, 1024)
		_, err = dial.Read(result)
		if err != nil {
			return
		}
		ending := findBytesEnding(result)
		var resp Response
		err := json.Unmarshal(result[:ending], &resp)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
