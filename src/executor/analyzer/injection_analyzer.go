package analyzer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
)

const PairName = "network_pairs"
const PairSplit = "->"

type InjectionAnalyzer struct {
	nsPrefixMap      map[string]struct{}
	faultResouceMap  map[string]chaos.ResourceField
	resourcesMaps    map[string]map[string][]string
	deduplicatedMaps map[string]map[string][]string
}

type countItem struct {
	faultType string
	node      *chaos.Node
	service   string
	isPair    bool
}

type coverageItem struct {
	num        int
	rangeNum   int
	coveredMap map[string]struct{}
}

func NewCoverageItem(rangeNum int, coveredMap map[string]struct{}) *coverageItem {
	return &coverageItem{
		num:        1,
		rangeNum:   rangeNum,
		coveredMap: coveredMap,
	}
}

func NewInjectionAnalyzer() (*InjectionAnalyzer, error) {
	nsPrefixMap := make(map[string]struct{}, len(config.GetNsPrefixs()))
	for _, nsPrefix := range config.GetNsPrefixs() {
		nsPrefixMap[nsPrefix] = struct{}{}
	}

	faultResourceMap, err := chaos.GetChaosResourceMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get fault resource map: %v", err)
	}

	chaosResourcesMap, err := chaos.GetNsResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get all resources: %v", err)
	}

	resourceMaps := make(map[string]map[string][]string)
	for nsPrefix, chaosResources := range chaosResourcesMap {
		resourceMaps[nsPrefix] = chaosResources.ToMap()
	}

	deduplicatedMaps := make(map[string]map[string][]string)
	for nsPrefix, chaosResources := range chaosResourcesMap {
		deduplicatedMaps[nsPrefix] = chaosResources.ToDeduplicatedMap()
	}

	return &InjectionAnalyzer{
		nsPrefixMap:      nsPrefixMap,
		faultResouceMap:  faultResourceMap,
		resourcesMaps:    resourceMaps,
		deduplicatedMaps: deduplicatedMaps,
	}, nil
}

func (a *InjectionAnalyzer) Analyze(injections []database.FaultInjectionProject) (map[string]dto.InjectionStats, error) {
	nodeMap := make(map[string][]chaos.Node, len(a.nsPrefixMap))
	for _, injection := range injections {
		node, nsPrefix, err := a.getNode(&injection)
		if err != nil {
			return nil, fmt.Errorf("failed to get node: %v", err)
		}

		if _, exists := nodeMap[nsPrefix]; !exists {
			nodeMap[nsPrefix] = []chaos.Node{*node}
		} else {
			nodeMap[nsPrefix] = append(nodeMap[nsPrefix], *node)
		}
	}

	statsMap := make(map[string]dto.InjectionStats, len(a.nsPrefixMap))
	for nsPrefix, nodes := range nodeMap {
		reference, err := chaos.StructToNode[chaos.InjectionConf](nsPrefix)
		if err != nil {
			return nil, fmt.Errorf("failed to get reference node for namespace prefix %s: %v", nsPrefix, err)
		}

		if reference == nil || len(reference.Children) == 0 {
			return nil, fmt.Errorf("invalid reference node: %v", reference)
		}

		var countItems []countItem
		for _, node := range nodes {
			faultTypeIndex := chaos.ChaosType(node.Value)
			faultType, exists := chaos.ChaosTypeMap[faultTypeIndex]
			if !exists {
				return nil, fmt.Errorf("invalid chaos type: %d", faultTypeIndex)
			}

			service, isPair, err := a.getService(&node, nsPrefix)
			if err != nil {
				return nil, fmt.Errorf("failed to get service for node %v: %v", node, err)
			}

			countItems = append(countItems, countItem{
				faultType: faultType,
				node:      &node,
				service:   service,
				isPair:    isPair,
			})
		}

		diversity, err := a.getDistributions(countItems, reference, nsPrefix)
		if err != nil {
			return nil, fmt.Errorf("failed to get injection diversity of namespace prefix %s: %v", nsPrefix, err)
		}

		statsMap[nsPrefix] = dto.InjectionStats{
			Diversity: *diversity,
		}
	}

	return statsMap, nil
}

func (a *InjectionAnalyzer) getDistributions(items []countItem, reference *chaos.Node, nsPrefix string) (*dto.InjectionDiversity, error) {
	faultMap := make(map[string]int)
	serviceMap := make(map[string]int)
	pairMap := make(map[string]*dto.PairStats)

	faultItemMap := make(map[string]countItem)
	faultServiceMap := make(map[string]map[string]struct{})

	for _, item := range items {
		faultMap[item.faultType]++
		faultItemMap[item.faultType] = item

		if _, faultExists := faultServiceMap[item.faultType]; !faultExists {
			faultServiceMap[item.faultType] = make(map[string]struct{})
			faultServiceMap[item.faultType][item.service] = struct{}{}
		} else {
			if _, serviceExists := faultServiceMap[item.faultType][item.service]; !serviceExists {
				faultServiceMap[item.faultType][item.service] = struct{}{}
			}
		}

		if !item.isPair {
			serviceMap[item.service]++
		} else {
			source, target, err := extractPairService(item.service)
			if err != nil {
				return nil, fmt.Errorf("failed to extract pair service from %s: %v", item.service, err)
			}

			if stats, exists := pairMap[source]; exists {
				stats.OutDegree++
			} else {
				pairMap[source] = &dto.PairStats{
					Name:      source,
					OutDegree: 1,
				}
			}

			if stats, exists := pairMap[target]; exists {
				stats.InDegree++
			} else {
				pairMap[target] = &dto.PairStats{
					Name:     target,
					InDegree: 1,
				}
			}
		}
	}

	pairStatsList := make([]dto.PairStats, 0, len(pairMap))
	for _, stats := range pairMap {
		pairStatsList = append(pairStatsList, *stats)
	}

	serviceCoverages := make(map[string]dto.ServiceCoverageItem, len(faultServiceMap))
	for faultType, serviceMap := range faultServiceMap {
		resourceName := a.faultResouceMap[faultType].Name
		resource := a.deduplicatedMaps[nsPrefix][resourceName]

		ratio := 0.0
		if len(resource) != 0 {
			ratio = float64(len(serviceMap)) / float64(len(resource))
		}

		var notCovered []string
		for _, service := range resource {
			if _, exists := serviceMap[service]; !exists {
				notCovered = append(notCovered, service)
			}
		}

		serviceCoverages[faultType] = dto.ServiceCoverageItem{
			Num:        len(serviceMap),
			NotCovered: notCovered,
			Coverage:   ratio,
		}
	}

	faultRangeNumMap := make(map[string]int, len(reference.Children))
	for key, node := range reference.Children {
		rangeNum, err := recursiveToGetRangeNum(node, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get attribute range num: %v", err)
		}

		faultRangeNumMap[node.Name] = rangeNum
	}

	coverageItemMap := make(map[string]map[string]*coverageItem, len(faultServiceMap))
	for _, item := range items {
		coveredMap, err := recursiveToGetCoveredMap(item.node, strconv.Itoa(item.node.Value))
		if err != nil {
			return nil, fmt.Errorf("failed to get covered map: %v", err)
		}

		_, faultExists := coverageItemMap[item.faultType]
		if !faultExists {
			coverageItemMap[item.faultType] = make(map[string]*coverageItem, len(faultServiceMap[item.faultType]))
			coverageItemMap[item.faultType][item.service] = NewCoverageItem(faultRangeNumMap[item.faultType], coveredMap)
			continue
		}

		_, serviceExists := coverageItemMap[item.faultType][item.service]
		if !serviceExists {
			coverageItemMap[item.faultType][item.service] = NewCoverageItem(faultRangeNumMap[item.faultType], coveredMap)
			continue
		}

		coverage := coverageItemMap[item.faultType][item.service]
		coverage.num += 1
		for v := range coveredMap {
			coverage.coveredMap[v] = struct{}{}
		}
	}

	attributeCoverages := make(map[string]map[string]dto.AttributeCoverageItem, len(coverageItemMap))
	for faultType, serviceMap := range coverageItemMap {
		attributeCoverages[faultType] = make(map[string]dto.AttributeCoverageItem, len(serviceMap))
		for service, coverageItem := range serviceMap {
			ratio := 0.0
			if coverageItem.rangeNum != 0 {
				ratio = float64(len(coverageItem.coveredMap)) / float64(coverageItem.rangeNum)
			}

			attributeCoverages[faultType][service] = dto.AttributeCoverageItem{
				Num:      coverageItem.num,
				Coverage: ratio,
			}
		}
	}

	return &dto.InjectionDiversity{
		FaultDistribution:   faultMap,
		ServiceDistribution: serviceMap,
		PairDistribution:    pairStatsList,
		ServiceCoverages:    serviceCoverages,
		AttributeCoverages:  attributeCoverages,
	}, nil
}

func (a *InjectionAnalyzer) getNode(item *database.FaultInjectionProject) (*chaos.Node, string, error) {
	nsPrefix, err := extractNamespacePrefix(item.InjectionName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract namespace prefix: %v", err)
	}

	if _, exists := a.nsPrefixMap[nsPrefix]; !exists {
		return nil, "", fmt.Errorf("namespace prefix %s not found in config", nsPrefix)
	}

	var node chaos.Node
	if err := json.Unmarshal([]byte(item.EngineConfig), &node); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal engine config: %v", err)
	}

	if _, err := chaos.Validate[chaos.InjectionConf](&node, nsPrefix); err != nil {
		return nil, "", fmt.Errorf("invalid engine config: %v", err)
	}

	return &node, nsPrefix, nil
}

func (a *InjectionAnalyzer) getService(node *chaos.Node, nsPrefix string) (string, bool, error) {
	faultTypeIndex := chaos.ChaosTypeMap[chaos.ChaosType(node.Value)]
	faultResourceName := a.faultResouceMap[faultTypeIndex].Name
	faultResource := a.resourcesMaps[nsPrefix][faultResourceName]

	child := node.Children[strconv.Itoa(node.Value)]
	serviceIndex := child.Children["2"].Value

	if serviceIndex >= len(faultResource) {
		return "", false, fmt.Errorf("service index out of bounds of %s: %d", faultResourceName, serviceIndex)
	}

	return faultResource[serviceIndex], faultResourceName == PairName, nil
}

func extractNamespacePrefix(name string) (string, error) {
	re := regexp.MustCompile(`^([a-zA-Z]+)\d+-([a-zA-Z]+)`)
	matches := re.FindStringSubmatch(name)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid name format: %s", name)
	}

	return matches[1], nil
}

func extractPairService(pairStr string) (string, string, error) {
	parts := strings.Split(pairStr, PairSplit)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid service pair format, expected 'service1->service2': %s", pairStr)
	}

	service1 := strings.TrimSpace(parts[0])
	service2 := strings.TrimSpace(parts[1])

	return service1, service2, nil
}

func recursiveToGetRangeNum(node *chaos.Node, key string) (int, error) {
	if node == nil {
		return 0, fmt.Errorf("invalid node: %v", node)
	}

	intKey, err := strconv.Atoi(key)
	if err != nil {
		return 0, fmt.Errorf("failed to convert key to int: %v", err)
	}

	total := 0
	if len(node.Children) == 0 && intKey > 2 {
		total = node.Range[1] - node.Range[0] + 1
	}

	if len(node.Children) != 0 {
		for childKey, childNode := range node.Children {
			num, err := recursiveToGetRangeNum(childNode, childKey)
			if err != nil {
				return 0, fmt.Errorf("failed to get range num for child node %s: %v", childKey, err)
			}

			total += num
		}
	}

	return total, nil
}

func recursiveToGetCoveredMap(node *chaos.Node, key string) (map[string]struct{}, error) {
	if node == nil {
		return nil, fmt.Errorf("invalid node: %v", node)
	}

	intKey, err := strconv.Atoi(key)
	if err != nil {
		return nil, fmt.Errorf("failed to convert key to int: %v", err)
	}

	covered := make(map[string]struct{})
	if len(node.Children) == 0 && intKey > 2 {
		covered[fmt.Sprintf("%s-%d", key, node.Value)] = struct{}{}
	}

	if len(node.Children) != 0 {
		for childKey, childNode := range node.Children {
			childCovered, err := recursiveToGetCoveredMap(childNode, childKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get covered map for child node %s: %v", childKey, err)
			}

			for v := range childCovered {
				covered[v] = struct{}{}
			}
		}
	}

	return covered, nil
}
