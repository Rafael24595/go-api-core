package swr

import (
	"github.com/Rafael24595/go-collections/collection"
)

type Step struct {
	Order int      `json:"order"`
	Type  StepType `json:"type"`
	Value string   `json:"value"`
}

func NewConditionStep(typ StepType, value string) *Step {
	return &Step{
		Order: 0,
		Type:  typ,
		Value: value,
	}
}

func OrderSteps(steps []Step) []Step {
	return collection.VectorFromList(steps).Sort(func(i, j Step) bool {
		return i.Order < j.Order
	}).Collect()
}

func FixStepsOrder(steps []Step) []Step {
	for i := range steps {
		steps[i].Order = i
	}
	return steps
}
