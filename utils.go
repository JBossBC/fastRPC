package fastRPC

type DataStandards struct {
	//open-tracing use
	//including compression ways
	header map[string]any
	code   int
	data   interface{}
}
