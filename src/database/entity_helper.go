package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

// Groundtruth represents the expected impact of a fault injection or execution
type Groundtruth struct {
	Service   []string `json:"service,omitempty"`
	Pod       []string `json:"pod,omitempty"`
	Container []string `json:"container,omitempty"`
	Metric    []string `json:"metric,omitempty"`
	Function  []string `json:"function,omitempty"`
	Span      []string `json:"span,omitempty"`
}

func NewDBGroundtruth(gt *chaos.Groundtruth) *Groundtruth {
	if gt == nil {
		return nil
	}

	return &Groundtruth{
		Service:   gt.Service,
		Pod:       gt.Pod,
		Container: gt.Container,
		Metric:    gt.Metric,
		Function:  gt.Function,
		Span:      gt.Span,
	}
}

// Scan implements sql.Scanner interface for reading from database
func (g *Groundtruth) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal Groundtruth value")
	}

	return json.Unmarshal(bytes, g)
}

// Value implements driver.Valuer interface for writing to database
func (g *Groundtruth) Value() (driver.Value, error) {
	if len(g.Service) == 0 && len(g.Pod) == 0 && len(g.Container) == 0 &&
		len(g.Metric) == 0 && len(g.Function) == 0 && len(g.Span) == 0 {
		return nil, nil
	}
	return json.Marshal(g)
}

func (g *Groundtruth) ConvertToChaosGroundtruth() *chaos.Groundtruth {
	if g == nil {
		return nil
	}

	return &chaos.Groundtruth{
		Service:   g.Service,
		Pod:       g.Pod,
		Container: g.Container,
		Metric:    g.Metric,
		Function:  g.Function,
		Span:      g.Span,
	}
}
