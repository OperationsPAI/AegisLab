package dto

type AlgorithmListResp struct {
	Algorithms []string `json:"algorithms"`
	Benchmarks []string `json:"benchmarks"`
}

type AlgorithmExecutionPayload struct {
	Benchmark   string `json:"benchmark"`
	Algorithm   string `json:"algorithm"`
	DatasetName string `json:"dataset"`
	Service     string `json:"service"`
	Tag         string `json:"tag"`
}
