package cmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	aegisclient "aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"
	"aegis/dto"
	"aegis/service/producer"

	chaoshandler "github.com/OperationsPAI/chaos-experiment/handler"
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
	client             *aegisclient.Client
	resolveProject     func() (int, error)
	resolveSystemIndex func(ctx context.Context, system string) (int, error)
	defaults           *chaosDefaults
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

	body, err := b.buildSubmission(context.Background(), spec)
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

func (b *backendSubmitter) buildSubmission(ctx context.Context, spec chaoscli.Spec) (map[string]any, error) {
	friendly := dto.FriendlyFaultSpec{
		Type:      spec.Type,
		Namespace: spec.Namespace,
		Target:    spec.Target,
		Duration:  spec.Duration,
		Params:    toProducerParams(spec),
	}

	resolvedTarget, err := resolveSpecTarget(ctx, spec)
	if err != nil {
		return nil, err
	}
	friendly.Target = strconv.Itoa(resolvedTarget)

	node, err := producer.FriendlySpecToNode(&friendly)
	if err != nil {
		return nil, err
	}
	systemIdx, err := b.getSystemIndex(ctx, spec.Namespace)
	if err != nil {
		return nil, err
	}
	applySystemIndex(&node, systemIdx)

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

func resolveSpecTarget(ctx context.Context, spec chaoscli.Spec) (int, error) {
	if idx, err := strconv.Atoi(spec.Target); err == nil {
		return idx, nil
	}

	switch spec.Type {
	case "NetworkDelay", "NetworkLoss":
		targetService, _ := spec.Params["target_service"].(string)
		return chaoshandler.ResolveNetworkPairIndex(spec.Namespace, spec.Target, targetService)
	case "HTTPRequestAbort":
		route, _ := spec.Params["route"].(string)
		method, _ := spec.Params["http_method"].(string)
		port := intFrom(spec.Params, "port", 80)
		return chaoshandler.ResolveHTTPEndpointIndex(spec.Namespace, spec.Target, route, method, port)
	case "JVMLatency":
		className, _ := spec.Params["class"].(string)
		methodName, _ := spec.Params["method"].(string)
		return chaoshandler.ResolveJVMMethodIndex(spec.Namespace, spec.Target, className, methodName)
	case "CPUStress":
		container, _ := spec.Params["container"].(string)
		return chaoshandler.ResolveContainerIndex(ctx, spec.Namespace, spec.Namespace, spec.Target, container)
	case "PodFailure":
		return chaoshandler.ResolveAppIndex(ctx, spec.Namespace, spec.Namespace, spec.Target)
	default:
		return 0, fmt.Errorf("unsupported chaos type %q", spec.Type)
	}
}

func (b *backendSubmitter) getSystemIndex(ctx context.Context, system string) (int, error) {
	if b.resolveSystemIndex != nil {
		return b.resolveSystemIndex(ctx, system)
	}
	if b.client == nil {
		b.client = newClient()
	}

	path := "/api/v2/injections/metadata?system=" + url.QueryEscape(system)
	var resp aegisclient.APIResponse[struct {
		SystemMap map[string]int `json:"system_map"`
	}]
	if err := b.client.Get(path, &resp); err != nil {
		return 0, err
	}

	for key, idx := range resp.Data.SystemMap {
		if key == system || strings.EqualFold(key, system) {
			return idx, nil
		}
	}

	return 0, fmt.Errorf("system %q not found in backend metadata", system)
}

func applySystemIndex(node *chaoshandler.Node, systemIdx int) {
	if node == nil {
		return
	}
	typeKey := strconv.Itoa(node.Value)
	typeNode := node.Children[typeKey]
	if typeNode == nil || typeNode.Children == nil {
		return
	}
	if systemNode := typeNode.Children["1"]; systemNode != nil {
		systemNode.Value = systemIdx
	}
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
