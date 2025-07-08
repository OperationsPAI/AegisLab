package dto

import "github.com/LGU-SE-Internal/rcabench/consts"

type GetContainerFilterOptions struct {
	Type  consts.ContainerType
	Name  string
	Image string
	Tag   string
}

type ListContainersFilterOptions struct {
	Status *bool
	Type   consts.ContainerType
	Names  []string
}

var ValidContainerTypes = map[consts.ContainerType]struct{}{
	consts.ContainerTypeAlgorithm: {},
	consts.ContainerTypeBenchmark: {},
}
