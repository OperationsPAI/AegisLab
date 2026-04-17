package cmd

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"aegis/cmd/aegisctl/client"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

type injectionMetadata struct {
	Config         *chaos.Node          `json:"config"`
	SystemResource chaos.SystemResource `json:"ns_resources"`
}

type metadataFetcher func(system string) (*injectionMetadata, error)

func fetchInjectionMetadata(system string) (*injectionMetadata, error) {
	c := newClient()
	path := "/api/v2/injections/metadata?system=" + url.QueryEscape(system)

	var resp client.APIResponse[injectionMetadata]
	if err := c.Get(path, &resp); err != nil {
		return nil, fmt.Errorf("fetch injection metadata for pedestal %q: %w", system, err)
	}
	if resp.Data.Config == nil {
		return nil, fmt.Errorf("injection metadata for pedestal %q did not include config", system)
	}

	return &resp.Data, nil
}

func translateInjectSpecFile(fileSpec InjectSpecFile, fetcher metadataFetcher) (*InjectSpec, error) {
	result := &InjectSpec{
		Pedestal:    fileSpec.Pedestal,
		Benchmark:   fileSpec.Benchmark,
		Interval:    fileSpec.Interval,
		PreDuration: fileSpec.PreDuration,
		Algorithms:  fileSpec.Algorithms,
		Labels:      fileSpec.Labels,
		Specs:       make([][]chaos.Node, 0, len(fileSpec.Specs)),
	}

	var md *injectionMetadata
	for batchIdx, batch := range fileSpec.Specs {
		outBatch := make([]chaos.Node, 0, len(batch))
		for specIdx, spec := range batch {
			if spec.isRawDSL() {
				node, err := spec.toRawNode()
				if err != nil {
					return nil, fmt.Errorf("invalid raw DSL in batch %d spec %d: %w", batchIdx, specIdx, err)
				}
				outBatch = append(outBatch, node)
				continue
			}

			if md == nil {
				var err error
				md, err = fetcher(fileSpec.Pedestal.Name)
				if err != nil {
					return nil, err
				}
			}

			node, err := translateHumanFaultSpec(spec, md)
			if err != nil {
				return nil, fmt.Errorf("translate fault spec in batch %d spec %d: %w", batchIdx, specIdx, err)
			}
			outBatch = append(outBatch, node)
		}
		result.Specs = append(result.Specs, outBatch)
	}

	return result, nil
}

func (f FaultSpec) isRawDSL() bool {
	return f.Value != nil || len(f.Children) > 0
}

func (f FaultSpec) toRawNode() (chaos.Node, error) {
	node := chaos.Node{
		Children: cloneNodeMap(f.Children),
		Value:    chaos.ValueNotSet,
	}
	if f.Value != nil {
		node.Value = *f.Value
	}
	if node.Value == chaos.ValueNotSet && len(node.Children) == 0 {
		return chaos.Node{}, fmt.Errorf("expected value and/or children")
	}
	return node, nil
}

func translateHumanFaultSpec(spec FaultSpec, md *injectionMetadata) (chaos.Node, error) {
	if md == nil || md.Config == nil {
		return chaos.Node{}, fmt.Errorf("missing injection metadata")
	}

	typeKey, typeNode, err := resolveFaultTypeNode(spec.Type, md.Config)
	if err != nil {
		return chaos.Node{}, err
	}

	root := chaos.Node{
		Value:    mustAtoi(typeKey),
		Children: map[string]*chaos.Node{typeKey: cloneNode(typeNode)},
	}
	active := root.Children[typeKey]

	fieldNodes := indexFieldNodes(active)

	if node, ok := fieldNodes["Duration"]; ok {
		duration, err := parseDurationMinutes(spec.Duration)
		if err != nil {
			return chaos.Node{}, fmt.Errorf("parse duration %q: %w", spec.Duration, err)
		}
		node.Value = duration
	}

	if node, ok := findResourceField(fieldNodes); ok {
		resourceName, resources, err := resourceChoices(node.Name, md)
		if err != nil {
			return chaos.Node{}, err
		}
		if strings.TrimSpace(spec.Target) == "" {
			return chaos.Node{}, fmt.Errorf("target is required for %s", resourceName)
		}
		idx, err := resolveTargetIndex(spec.Target, resources)
		if err != nil {
			return chaos.Node{}, fmt.Errorf("resolve target %q in %s: %w", spec.Target, resourceName, err)
		}
		node.Value = idx
	}

	extras := normalizedExtraMap(spec)
	for _, key := range orderedFieldKeys(fieldNodes) {
		node := fieldNodes[key]
		if node.Name == "Duration" || node.Name == "Namespace" || node.Name == "NamespaceTarget" || isResourceField(node.Name) {
			continue
		}

		if value, ok := extras[normalizeToken(node.Name)]; ok {
			parsed, err := parseIntValue(value)
			if err != nil {
				return chaos.Node{}, fmt.Errorf("parse field %s: %w", node.Name, err)
			}
			node.Value = parsed
			continue
		}

		if node.Value == chaos.ValueNotSet {
			node.Value = defaultNodeValue(node)
		}
	}

	if node, ok := fieldNodes["NamespaceTarget"]; ok && node.Value == chaos.ValueNotSet {
		node.Value = defaultNodeValue(node)
	}

	return root, nil
}

func resolveFaultTypeNode(typeName string, config *chaos.Node) (string, *chaos.Node, error) {
	if config == nil || len(config.Children) == 0 {
		return "", nil, fmt.Errorf("metadata config is empty")
	}

	want := normalizeToken(typeName)
	for key, child := range config.Children {
		if normalizeToken(child.Name) == want {
			return key, child, nil
		}
	}

	available := make([]string, 0, len(config.Children))
	for _, child := range config.Children {
		available = append(available, child.Name)
	}
	return "", nil, fmt.Errorf("unknown fault type %q (available: %s)", typeName, strings.Join(available, ", "))
}

func indexFieldNodes(node *chaos.Node) map[string]*chaos.Node {
	out := make(map[string]*chaos.Node, len(node.Children))
	for _, child := range node.Children {
		out[child.Name] = child
	}
	return out
}

func orderedFieldKeys(fields map[string]*chaos.Node) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	return keys
}

func findResourceField(fields map[string]*chaos.Node) (*chaos.Node, bool) {
	for _, name := range []string{"AppIdx", "ContainerIdx", "EndpointIdx", "DNSEndpointIdx"} {
		if node, ok := fields[name]; ok {
			return node, true
		}
	}
	return nil, false
}

func isResourceField(name string) bool {
	switch name {
	case "AppIdx", "ContainerIdx", "EndpointIdx", "DNSEndpointIdx":
		return true
	default:
		return false
	}
}

func resourceChoices(fieldName string, md *injectionMetadata) (string, []string, error) {
	switch fieldName {
	case "AppIdx":
		if len(md.SystemResource.Services) == 0 {
			return "services", nil, fmt.Errorf("metadata did not include services")
		}
		return "services", md.SystemResource.Services, nil
	case "ContainerIdx":
		if len(md.SystemResource.Containers) > 0 {
			return "containers", md.SystemResource.Containers, nil
		}
		if len(md.SystemResource.Services) > 0 {
			return "services", md.SystemResource.Services, nil
		}
		return "containers", nil, fmt.Errorf("metadata did not include containers or services")
	case "EndpointIdx", "DNSEndpointIdx":
		if len(md.SystemResource.Endpoints) == 0 {
			return "endpoints", nil, fmt.Errorf("metadata did not include endpoints")
		}
		return "endpoints", md.SystemResource.Endpoints, nil
	default:
		return fieldName, nil, fmt.Errorf("fault field %s is not supported by the human-readable translator", fieldName)
	}
}

func resolveTargetIndex(target string, resources []string) (int, error) {
	if len(resources) == 0 {
		return 0, fmt.Errorf("no resources available")
	}

	targetNorm := normalizeToken(target)

	exact := make([]int, 0, 1)
	contains := make([]int, 0, 1)
	for idx, resource := range resources {
		resourceNorm := normalizeToken(resource)
		if resourceNorm == targetNorm {
			exact = append(exact, idx)
			continue
		}
		if strings.Contains(resourceNorm, targetNorm) || strings.Contains(targetNorm, resourceNorm) {
			contains = append(contains, idx)
		}
	}

	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) > 1 {
		return 0, fmt.Errorf("target matched multiple resources exactly")
	}
	if len(contains) == 1 {
		return contains[0], nil
	}
	if len(contains) > 1 {
		return 0, fmt.Errorf("target matched multiple resources")
	}

	return 0, fmt.Errorf("target not found (available: %s)", strings.Join(resources, ", "))
}

func normalizedExtraMap(spec FaultSpec) map[string]any {
	out := make(map[string]any, len(spec.Extra))
	for key, value := range spec.Extra {
		out[normalizeToken(key)] = value
	}
	return out
}

func parseDurationMinutes(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("duration is required")
	}

	if n, err := strconv.Atoi(raw); err == nil {
		if n <= 0 {
			return 0, fmt.Errorf("duration must be > 0")
		}
		return n, nil
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be > 0")
	}

	return int(math.Ceil(d.Minutes())), nil
}

func parseIntValue(v any) (int, error) {
	switch value := v.(type) {
	case int:
		return value, nil
	case int64:
		return int(value), nil
	case float64:
		return int(value), nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, err
		}
		return n, nil
	default:
		return 0, fmt.Errorf("unsupported value type %T", v)
	}
}

func defaultNodeValue(node *chaos.Node) int {
	if node == nil {
		return 0
	}
	if node.Value != chaos.ValueNotSet {
		return node.Value
	}
	if len(node.Range) >= 1 {
		return node.Range[0]
	}
	return 0
}

func cloneNode(node *chaos.Node) *chaos.Node {
	if node == nil {
		return nil
	}
	cloned := *node
	cloned.Children = cloneNodeMap(node.Children)
	if node.Range != nil {
		cloned.Range = append([]int(nil), node.Range...)
	}
	return &cloned
}

func cloneNodeMap(children map[string]*chaos.Node) map[string]*chaos.Node {
	if len(children) == 0 {
		return nil
	}
	out := make(map[string]*chaos.Node, len(children))
	for key, child := range children {
		out[key] = cloneNode(child)
	}
	return out
}

func normalizeToken(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}
