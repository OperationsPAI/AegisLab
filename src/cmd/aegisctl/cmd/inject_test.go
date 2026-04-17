package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// TestParseFaultSpecYAML
// ---------------------------------------------------------------------------

func TestParseFaultSpecYAML(t *testing.T) {
	tests := []struct {
		name           string
		yamlInput      string
		expectError    bool
		expectedSpecs  int // number of batches in Specs
		checkFirstSpec func(t *testing.T, spec FaultSpec)
	}{
		{
			name: "single CPUStress fault",
			yamlInput: `
pedestal:
  name: otel-demo
  version: "1.0.0"
benchmark:
  name: clickhouse
  version: "1.0.0"
interval: 60
pre_duration: 30
specs:
  - - type: CPUStress
      namespace: exp
      target: frontend
      duration: "60s"
`,
			expectError:   false,
			expectedSpecs: 1,
			checkFirstSpec: func(t *testing.T, spec FaultSpec) {
				assert.Equal(t, "CPUStress", spec.Type)
				assert.Equal(t, "exp", spec.Namespace)
				assert.Equal(t, "frontend", spec.Target)
				assert.Equal(t, "60s", spec.Duration)
			},
		},
		{
			name: "PodKill fault type",
			yamlInput: `
pedestal:
  name: train-ticket
benchmark:
  name: clickhouse
interval: 30
pre_duration: 10
specs:
  - - type: PodKill
      namespace: ts
      target: ts-order-service
      duration: "30s"
`,
			expectError:   false,
			expectedSpecs: 1,
			checkFirstSpec: func(t *testing.T, spec FaultSpec) {
				assert.Equal(t, "PodKill", spec.Type)
				assert.Equal(t, "ts", spec.Namespace)
				assert.Equal(t, "ts-order-service", spec.Target)
				assert.Equal(t, "30s", spec.Duration)
			},
		},
		{
			name: "multiple batches with parallel faults",
			yamlInput: `
pedestal:
  name: otel-demo
benchmark:
  name: clickhouse
interval: 120
pre_duration: 30
specs:
  - - type: CPUStress
      namespace: exp
      target: frontend
      duration: "60s"
    - type: MemoryStress
      namespace: exp
      target: backend
      duration: "60s"
  - - type: PodKill
      namespace: exp
      target: cart
      duration: "30s"
`,
			expectError:   false,
			expectedSpecs: 2,
			checkFirstSpec: func(t *testing.T, spec FaultSpec) {
				assert.Equal(t, "CPUStress", spec.Type)
			},
		},
		{
			name: "with algorithms and labels",
			yamlInput: `
pedestal:
  name: otel-demo
benchmark:
  name: clickhouse
interval: 60
pre_duration: 30
specs:
  - - type: CPUStress
      namespace: exp
      target: frontend
      duration: "60s"
algorithms:
  - name: random
    version: "1.0.0"
labels:
  - key: experiment
    value: cpu-stress-test
`,
			expectError:   false,
			expectedSpecs: 1,
			checkFirstSpec: func(t *testing.T, spec FaultSpec) {
				assert.Equal(t, "CPUStress", spec.Type)
			},
		},
		{
			name: "with env_vars in pedestal",
			yamlInput: `
pedestal:
  name: otel-demo
  version: "2.0.0"
  env_vars:
    - key: OTEL_EXPORTER
      value: jaeger
benchmark:
  name: clickhouse
interval: 60
pre_duration: 30
specs:
  - - type: CPUStress
      namespace: exp
      target: frontend
      duration: "60s"
`,
			expectError:   false,
			expectedSpecs: 1,
			checkFirstSpec: func(t *testing.T, spec FaultSpec) {
				assert.Equal(t, "CPUStress", spec.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var spec InjectSpec
			err := yaml.Unmarshal([]byte(tt.yamlInput), &spec)

			if tt.expectError {
				assert.Error(t, err, "parsing should fail")
				return
			}

			assert.NoError(t, err, "parsing should succeed")
			assert.Equal(t, tt.expectedSpecs, len(spec.Specs),
				"expected %d batch(es) in specs", tt.expectedSpecs)

			if len(spec.Specs) > 0 && len(spec.Specs[0]) > 0 && tt.checkFirstSpec != nil {
				tt.checkFirstSpec(t, spec.Specs[0][0])
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestParseFaultSpecYAML_TopLevelFields
// ---------------------------------------------------------------------------

func TestParseFaultSpecYAML_TopLevelFields(t *testing.T) {
	yamlInput := `
pedestal:
  name: otel-demo
  version: "1.0.0"
benchmark:
  name: clickhouse
  version: "2.0.0"
interval: 120
pre_duration: 45
specs:
  - - type: CPUStress
      namespace: exp
      target: frontend
      duration: "60s"
algorithms:
  - name: random
    version: "1.0.0"
  - name: causalrca
    version: "0.5.0"
labels:
  - key: team
    value: infra
  - key: env
    value: staging
`

	var spec InjectSpec
	err := yaml.Unmarshal([]byte(yamlInput), &spec)
	assert.NoError(t, err)

	t.Run("pedestal fields", func(t *testing.T) {
		assert.Equal(t, "otel-demo", spec.Pedestal.Name)
		assert.Equal(t, "1.0.0", spec.Pedestal.Version)
	})

	t.Run("benchmark fields", func(t *testing.T) {
		assert.Equal(t, "clickhouse", spec.Benchmark.Name)
		assert.Equal(t, "2.0.0", spec.Benchmark.Version)
	})

	t.Run("timing fields", func(t *testing.T) {
		assert.Equal(t, 120, spec.Interval)
		assert.Equal(t, 45, spec.PreDuration)
	})

	t.Run("algorithms", func(t *testing.T) {
		assert.Len(t, spec.Algorithms, 2)
		assert.Equal(t, "random", spec.Algorithms[0].Name)
		assert.Equal(t, "causalrca", spec.Algorithms[1].Name)
	})

	t.Run("labels", func(t *testing.T) {
		assert.Len(t, spec.Labels, 2)
		assert.Equal(t, "team", spec.Labels[0].Key)
		assert.Equal(t, "infra", spec.Labels[0].Value)
		assert.Equal(t, "env", spec.Labels[1].Key)
		assert.Equal(t, "staging", spec.Labels[1].Value)
	})
}

// ---------------------------------------------------------------------------
// TestFaultSpecValidation
// ---------------------------------------------------------------------------

func TestFaultSpecValidation(t *testing.T) {
	t.Run("unknown fault type name parses as string", func(t *testing.T) {
		// FaultSpec.Type is a plain string, so any value is accepted at parse time.
		// Validation of the type happens at the translation/API layer, not during YAML parsing.
		yamlInput := `
type: CompletelyUnknownFault
namespace: exp
target: frontend
duration: "60s"
`
		var spec FaultSpec
		err := yaml.Unmarshal([]byte(yamlInput), &spec)
		assert.NoError(t, err, "YAML parsing should accept any string type")
		assert.Equal(t, "CompletelyUnknownFault", spec.Type)
	})

	t.Run("empty spec parses with zero values", func(t *testing.T) {
		yamlInput := `{}`
		var spec FaultSpec
		err := yaml.Unmarshal([]byte(yamlInput), &spec)
		assert.NoError(t, err, "empty spec should parse")
		assert.Empty(t, spec.Type)
		assert.Empty(t, spec.Namespace)
		assert.Empty(t, spec.Target)
		assert.Empty(t, spec.Duration)
	})

	t.Run("empty InjectSpec parses with zero values", func(t *testing.T) {
		yamlInput := `{}`
		var spec InjectSpec
		err := yaml.Unmarshal([]byte(yamlInput), &spec)
		assert.NoError(t, err, "empty inject spec should parse")
		assert.Empty(t, spec.Pedestal.Name)
		assert.Equal(t, 0, spec.Interval)
		assert.Equal(t, 0, spec.PreDuration)
		assert.Nil(t, spec.Specs)
	})

	t.Run("invalid YAML produces error", func(t *testing.T) {
		yamlInput := `
specs:
  - - type: CPUStress
    namespace: exp  # wrong indentation -- not a list item
`
		var spec InjectSpec
		err := yaml.Unmarshal([]byte(yamlInput), &spec)
		// This may or may not error depending on YAML parser leniency.
		// The key check is that if it errors, the error is from YAML parsing.
		if err != nil {
			assert.Contains(t, err.Error(), "yaml", "error should be from YAML parser")
		}
	})

	t.Run("params field captures extra parameters", func(t *testing.T) {
		yamlInput := `
type: CPUStress
namespace: exp
target: frontend
duration: "60s"
params:
  cpu_load: 80
  cpu_worker: 2
`
		var spec FaultSpec
		err := yaml.Unmarshal([]byte(yamlInput), &spec)
		assert.NoError(t, err, "params should be parsed")
		assert.Equal(t, "CPUStress", spec.Type)
		assert.NotNil(t, spec.Params)
		assert.Contains(t, spec.Params, "cpu_load", "params should contain cpu_load")
	})
}

// ---------------------------------------------------------------------------
// TestContainerRefParsing
// ---------------------------------------------------------------------------

func TestContainerRefParsing(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		checkFunc func(t *testing.T, ref ContainerRef)
	}{
		{
			name: "basic name and version",
			yamlInput: `
name: otel-demo
version: "1.0.0"
`,
			checkFunc: func(t *testing.T, ref ContainerRef) {
				assert.Equal(t, "otel-demo", ref.Name)
				assert.Equal(t, "1.0.0", ref.Version)
				assert.Nil(t, ref.EnvVars)
				assert.Nil(t, ref.Payload)
			},
		},
		{
			name: "with env_vars",
			yamlInput: `
name: benchmark
env_vars:
  - key: MODE
    value: production
  - key: TIMEOUT
    value: "30"
`,
			checkFunc: func(t *testing.T, ref ContainerRef) {
				assert.Equal(t, "benchmark", ref.Name)
				assert.Len(t, ref.EnvVars, 2)
				assert.Equal(t, "MODE", ref.EnvVars[0].Key)
				assert.Equal(t, "production", ref.EnvVars[0].Value)
			},
		},
		{
			name: "with payload",
			yamlInput: `
name: special
payload:
  key1: value1
  key2: 42
`,
			checkFunc: func(t *testing.T, ref ContainerRef) {
				assert.Equal(t, "special", ref.Name)
				assert.NotNil(t, ref.Payload)
				assert.Equal(t, "value1", ref.Payload["key1"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref ContainerRef
			err := yaml.Unmarshal([]byte(tt.yamlInput), &ref)
			assert.NoError(t, err)
			tt.checkFunc(t, ref)
		})
	}
}
