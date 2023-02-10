package fastRPC

type DataStandards struct {
	//open-tracing use
	//including compression ways
	Header map[string]any `json:"header"`
	Code   int            `json:"code"`
	Method string         `json:"method""`
	Data   []byte         `json:"data"`
}

type Response struct {
	Code   int            `json:"code"`
	Header map[string]any `json:"header"`
	Result []byte         `json:"result"`
}

func (resp *Response) GetFailedResp(message string) {
	resp.Result = []byte(message)
	resp.Code = 300
}
func (resp *Response) GetSuccessResp(data []byte) {
	resp.Result = data
	resp.Code = 200
}
func (resp *Response) SetHeader(header map[string]any) {
	for key, value := range header {
		resp.Header[key] = value
	}
}
func (resp *Response) GetHeader(key string) any {
	return resp.Header[key]
}
