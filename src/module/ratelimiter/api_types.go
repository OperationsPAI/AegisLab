package ratelimiter

// RateLimiterHolder describes a single task currently holding a token.
type RateLimiterHolder struct {
	TaskID     string `json:"task_id"`
	TaskState  string `json:"task_state"`
	IsTerminal bool   `json:"is_terminal"`
}

// RateLimiterItem describes one token-bucket rate limiter.
type RateLimiterItem struct {
	Bucket   string              `json:"bucket"`
	Key      string              `json:"key"`
	Capacity int                 `json:"capacity"`
	Held     int                 `json:"held"`
	Holders  []RateLimiterHolder `json:"holders"`
}

// RateLimiterListResp is the response for GET /api/v2/rate-limiters.
type RateLimiterListResp struct {
	Items []RateLimiterItem `json:"items"`
}

// RateLimiterGCResp is the response for POST /api/v2/rate-limiters/gc.
type RateLimiterGCResp struct {
	Released       int `json:"released"`
	TouchedBuckets int `json:"touched_buckets"`
}
