package dto

type AlgorithmListResp struct {
	Algorithms []string `json:"algorithms"`
}

type AlgorithmExecutionPayload struct {
	Algorithm   string `json:"algorithm"`
	DatasetName string `json:"dataset"`
	Service     string `json:"service"`
	Tag         string `json:"tag"`
}
