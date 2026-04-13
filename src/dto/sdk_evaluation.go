package dto

import "fmt"

// ListSDKEvaluationReq represents the request for listing SDK evaluation samples.
type ListSDKEvaluationReq struct {
	PaginationReq
	ExpID string `form:"exp_id"`
	Stage string `form:"stage"` // "init", "rollout", "judged"
}

func (req *ListSDKEvaluationReq) Validate() error {
	if req.Stage != "" {
		validStages := map[string]struct{}{
			"init":    {},
			"rollout": {},
			"judged":  {},
		}
		if _, ok := validStages[req.Stage]; !ok {
			return fmt.Errorf("invalid stage: %s (must be one of: init, rollout, judged)", req.Stage)
		}
	}
	return req.PaginationReq.Validate()
}

// SDKExperimentListResp represents the response for listing SDK experiments.
type SDKExperimentListResp struct {
	Experiments []string `json:"experiments"`
}

// ListSDKDatasetSampleReq represents the request for listing SDK dataset samples.
type ListSDKDatasetSampleReq struct {
	PaginationReq
	Dataset string `form:"dataset"`
}

func (req *ListSDKDatasetSampleReq) Validate() error {
	return req.PaginationReq.Validate()
}
