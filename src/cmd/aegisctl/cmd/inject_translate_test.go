package cmd

import (
	"testing"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

func TestTranslateInjectSpecFileCPUStress(t *testing.T) {
	md := &injectionMetadata{
		Config: &chaos.Node{
			Children: map[string]*chaos.Node{
				"4": {
					Name: "CPUStress",
					Children: map[string]*chaos.Node{
						"0": {Name: "Duration", Range: []int{1, 60}, Value: chaos.ValueNotSet},
						"1": {Name: "Namespace", Range: []int{0, 5}, Value: 3},
						"2": {Name: "ContainerIdx", Range: []int{0, 10}, Value: chaos.ValueNotSet},
						"3": {Name: "CPULoad", Range: []int{1, 100}, Value: chaos.ValueNotSet},
						"4": {Name: "CPUWorker", Range: []int{1, 3}, Value: chaos.ValueNotSet},
						"5": {Name: "NamespaceTarget", Range: []int{0, 0}, Value: chaos.ValueNotSet},
					},
				},
			},
		},
		SystemResource: chaos.SystemResource{
			Containers: []string{"checkoutservice", "frontend", "recommendationservice"},
		},
	}

	spec, err := translateInjectSpecFile(InjectSpecFile{
		Pedestal: ContainerRef{Name: "otel-demo"},
		Specs: [][]FaultSpec{{
			{
				Type:      "CPUStress",
				Namespace: "exp",
				Target:    "frontend",
				Duration:  "60s",
				Extra: map[string]any{
					"cpu_load": 80,
				},
			},
		}},
	}, func(system string) (*injectionMetadata, error) {
		if system != "otel-demo" {
			t.Fatalf("unexpected system %q", system)
		}
		return md, nil
	})
	if err != nil {
		t.Fatalf("translateInjectSpecFile() error = %v", err)
	}

	got := spec.Specs[0][0]
	child := got.Children["4"]
	if got.Value != 4 {
		t.Fatalf("root value = %d, want 4", got.Value)
	}
	if child.Children["0"].Value != 1 {
		t.Fatalf("duration value = %d, want 1", child.Children["0"].Value)
	}
	if child.Children["1"].Value != 3 {
		t.Fatalf("namespace value = %d, want 3", child.Children["1"].Value)
	}
	if child.Children["2"].Value != 1 {
		t.Fatalf("container index = %d, want 1", child.Children["2"].Value)
	}
	if child.Children["3"].Value != 80 {
		t.Fatalf("cpu load = %d, want 80", child.Children["3"].Value)
	}
	if child.Children["4"].Value != 1 {
		t.Fatalf("cpu worker default = %d, want 1", child.Children["4"].Value)
	}
	if child.Children["5"].Value != 0 {
		t.Fatalf("namespace target default = %d, want 0", child.Children["5"].Value)
	}
}

func TestTranslateInjectSpecFilePodKill(t *testing.T) {
	md := &injectionMetadata{
		Config: &chaos.Node{
			Children: map[string]*chaos.Node{
				"0": {
					Name: "PodKill",
					Children: map[string]*chaos.Node{
						"0": {Name: "Duration", Range: []int{1, 60}, Value: chaos.ValueNotSet},
						"1": {Name: "Namespace", Range: []int{0, 5}, Value: 5},
						"2": {Name: "AppIdx", Range: []int{0, 10}, Value: chaos.ValueNotSet},
						"3": {Name: "NamespaceTarget", Range: []int{0, 0}, Value: chaos.ValueNotSet},
					},
				},
			},
		},
		SystemResource: chaos.SystemResource{
			Services: []string{"ts-order-service", "ts-payment-service"},
		},
	}

	spec, err := translateInjectSpecFile(InjectSpecFile{
		Pedestal: ContainerRef{Name: "ts"},
		Specs: [][]FaultSpec{{
			{
				Type:      "pod-kill",
				Namespace: "ts",
				Target:    "ts-order-service",
				Duration:  "120s",
			},
		}},
	}, func(system string) (*injectionMetadata, error) {
		if system != "ts" {
			t.Fatalf("unexpected system %q", system)
		}
		return md, nil
	})
	if err != nil {
		t.Fatalf("translateInjectSpecFile() error = %v", err)
	}

	got := spec.Specs[0][0]
	child := got.Children["0"]
	if got.Value != 0 {
		t.Fatalf("root value = %d, want 0", got.Value)
	}
	if child.Children["0"].Value != 2 {
		t.Fatalf("duration value = %d, want 2", child.Children["0"].Value)
	}
	if child.Children["1"].Value != 5 {
		t.Fatalf("namespace value = %d, want 5", child.Children["1"].Value)
	}
	if child.Children["2"].Value != 0 {
		t.Fatalf("app index = %d, want 0", child.Children["2"].Value)
	}
	if child.Children["3"].Value != 0 {
		t.Fatalf("namespace target default = %d, want 0", child.Children["3"].Value)
	}
}
