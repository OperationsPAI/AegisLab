package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"

	aegisclient "aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"
	"aegis/dto"
	"aegis/service/producer"

	chaoscli "github.com/OperationsPAI/chaos-experiment/pkg/chaoscli"
	"gopkg.in/yaml.v3"
)

type chaosDefaults struct {
	pedestal    string
	benchmark   string
	interval    int
	preDuration int
}

type backendSubmitter struct {
	client         *aegisclient.Client
	resolveProject func() (int, error)
	defaults       *chaosDefaults
}

func newBackendSubmitter(defaults *chaosDefaults) *backendSubmitter {
	return &backendSubmitter{
		resolveProject: resolveProjectIDByName,
		defaults:       defaults,
	}
}

func (b *backendSubmitter) Submit(_ context.Context, spec chaoscli.Spec) error {
	if b.client == nil {
		b.client = newClient()
	}
	pid, err := b.resolveProject()
	if err != nil {
		return err
	}

	body, err := b.buildSubmission(spec)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/v2/projects/%d/injections/inject", pid)

	var resp aegisclient.APIResponse[any]
	if err := b.client.Post(path, body, &resp); err != nil {
		return err
	}
	output.PrintJSON(resp.Data)
	return nil
}

func (b *backendSubmitter) DryRun(_ context.Context, spec chaoscli.Spec, w io.Writer) error {
	data, err := yaml.Marshal(b.buildInjectSpec(spec))
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (b *backendSubmitter) buildInjectSpec(spec chaoscli.Spec) InjectSpec {
	return InjectSpec{
		Pedestal: ContainerRef{
			Name: b.defaults.pedestal,
		},
		Benchmark: ContainerRef{
			Name: b.defaults.benchmark,
		},
		Interval:    b.defaults.interval,
		PreDuration: b.defaults.preDuration,
		Specs: [][]FaultSpec{{
			{
				Type:      spec.Type,
				Namespace: spec.Namespace,
				Target:    spec.Target,
				Duration:  spec.Duration,
				Params:    spec.Params,
			},
		}},
	}
}

func (b *backendSubmitter) buildSubmission(spec chaoscli.Spec) (map[string]any, error) {
	friendly := dto.FriendlyFaultSpec{
		Type:      spec.Type,
		Namespace: spec.Namespace,
		Target:    spec.Target,
		Duration:  spec.Duration,
		Params:    toProducerParams(spec),
	}
	node, err := producer.FriendlySpecToNode(&friendly)
	if err != nil {
		return nil, err
	}

	injectSpec := b.buildInjectSpec(spec)
	return map[string]any{
		"pedestal":     injectSpec.Pedestal,
		"benchmark":    injectSpec.Benchmark,
		"interval":     injectSpec.Interval,
		"pre_duration": injectSpec.PreDuration,
		"specs": [][]any{{
			node,
		}},
	}, nil
}

func toProducerParams(spec chaoscli.Spec) map[string]any {
	params := map[string]any{}
	switch spec.Type {
	case "NetworkDelay":
		params["Latency"] = intFrom(spec.Params, "latency", 100)
		params["Correlation"] = intFrom(spec.Params, "correlation", 0)
		params["Jitter"] = intFrom(spec.Params, "jitter", 0)
		params["Direction"] = directionCode(spec.Params["direction"])
	case "NetworkLoss":
		params["Loss"] = intFrom(spec.Params, "loss", 10)
		params["Correlation"] = intFrom(spec.Params, "correlation", 0)
		params["Direction"] = directionCode(spec.Params["direction"])
	case "JVMLatency":
		params["LatencyDuration"] = intFrom(spec.Params, "latency_duration", 1000)
	case "CPUStress":
		params["CPULoad"] = intFrom(spec.Params, "cpu_load", 80)
		params["CPUWorker"] = intFrom(spec.Params, "cpu_worker", 1)
	}
	if len(params) == 0 {
		return nil
	}
	return params
}

func intFrom(params map[string]any, key string, fallback int) int {
	if params == nil {
		return fallback
	}
	switch v := params[key].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

func directionCode(v any) int {
	s, _ := v.(string)
	switch s {
	case "from":
		return 2
	case "both":
		return 3
	default:
		return 1
	}
}

func init() {
	defaults := &chaosDefaults{
		pedestal:    "ts",
		benchmark:   "clickhouse",
		interval:    60,
		preDuration: 30,
	}

	chaosCmd := chaoscli.NewRootCmd(newBackendSubmitter(defaults))
	chaosCmd.Short = "Create tracked chaos injections through the AegisLab backend"
	chaosCmd.PersistentFlags().StringVar(&defaults.pedestal, "pedestal", defaults.pedestal, "Pedestal container name")
	chaosCmd.PersistentFlags().StringVar(&defaults.benchmark, "benchmark", defaults.benchmark, "Benchmark container name")
	chaosCmd.PersistentFlags().IntVar(&defaults.interval, "interval", defaults.interval, "Experiment interval in minutes")
	chaosCmd.PersistentFlags().IntVar(&defaults.preDuration, "pre-duration", defaults.preDuration, "Pre-injection duration in minutes")

	rootCmd.AddCommand(chaosCmd)
}
