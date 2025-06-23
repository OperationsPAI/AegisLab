package dto

import "github.com/LGU-SE-Internal/rcabench/consts"

type FilterContainerOptions struct {
	Status *bool
	Type   consts.ContainerType
	Names  []string
}
