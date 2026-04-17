package cmd

import (
	"context"
	"io"
	"strconv"

	aegisclient "aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

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

	body := b.buildInjectSpec(spec)
	submission := translateSpecsIfPossible(b.client, body)
	path := "/api/v2/projects/" + strconv.Itoa(pid) + "/injections/inject"

	var resp aegisclient.APIResponse[any]
	if err := b.client.Post(path, submission, &resp); err != nil {
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

func (b *backendSubmitter) buildInjectSpec(spec chaoscli.Spec) *InjectSpec {
	return &InjectSpec{
		Pedestal: ContainerRef{
			Name: valueOrDefault(b.defaults.pedestal, spec.Namespace),
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

func valueOrDefault(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func init() {
	defaults := &chaosDefaults{
		pedestal:    "",
		benchmark:   "clickhouse",
		interval:    60,
		preDuration: 30,
	}

	chaosCmd := chaoscli.NewRootCmd(newBackendSubmitter(defaults))
	chaosCmd.Short = "Create tracked chaos injections through the AegisLab backend"
	chaosCmd.PersistentFlags().StringVar(&defaults.pedestal, "pedestal", defaults.pedestal, "Pedestal container name (defaults to --namespace)")
	chaosCmd.PersistentFlags().StringVar(&defaults.benchmark, "benchmark", defaults.benchmark, "Benchmark container name")
	chaosCmd.PersistentFlags().IntVar(&defaults.interval, "interval", defaults.interval, "Experiment interval in minutes")
	chaosCmd.PersistentFlags().IntVar(&defaults.preDuration, "pre-duration", defaults.preDuration, "Pre-injection duration in minutes")

	rootCmd.AddCommand(chaosCmd)
}
