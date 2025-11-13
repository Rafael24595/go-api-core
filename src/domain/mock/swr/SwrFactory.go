package swr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-collections/collection"
)

type UnmarshalOpts struct {
	Evalue bool
}

func Unmarshal(cond string) ([]mock.ConditionStep, error) {
	return unmarshal(cond, UnmarshalOpts{
		Evalue: false,
	})
}

func UnmarshalWithOptions(cond string, opts UnmarshalOpts) ([]mock.ConditionStep, error) {
	return unmarshal(cond, opts)
}

func unmarshal(cond string, opts UnmarshalOpts) ([]mock.ConditionStep, error) {
	fragments := collection.VectorFromList(utils.SplitByRune(cond, '.'))

	steps := make([]mock.ConditionStep, 0)

	headers, err := findHeaders(fragments)
	if err != nil {
		return steps, err
	}

	steps = append(steps, headers...)

	for fragments.Size() > 0 {
		cursor, ok := fragments.Shift()
		if !ok {
			break
		}

		var step *mock.ConditionStep
		if _, value, ok := strings.Cut(*cursor, "$"); ok {
			step = mock.NewConditionStep(mock.StepTypeOperator, value)
		} else if value, ok := findPosition(*cursor); ok {
			step = mock.NewConditionStep(mock.StepTypeArray, value)
		} else if value, ok := findValue(*cursor); ok {
			step = mock.NewConditionStep(mock.StepTypeValue, value)
		} else {
			step = mock.NewConditionStep(mock.StepTypeField, *cursor)
		}

		steps = append(steps, *step)
	}

	if opts.Evalue {
		if err := evalueTypes(steps); err != nil {
			return steps, err
		}
	}

	return steps, nil
}

func findHeaders(fragments *collection.Vector[string]) ([]mock.ConditionStep, error) {
	steps := make([]mock.ConditionStep, 0)
	target, ok := fragments.Shift()
	if !ok {
		return steps, nil
	}

	step := mock.NewConditionStep(mock.StepTypeInput, *target)
	steps = append(steps, *step)

	if *target != string(mock.StepInputPayload) {
		return steps, nil
	}

	format, ok := fragments.Shift()
	if !ok {
		return steps, errors.New("payload format is undefined")
	}

	step = mock.NewConditionStep(mock.StepTypeFormat, *format)
	steps = append(steps, *step)

	return steps, nil
}

func evalueTypes(conditions []mock.ConditionStep) error {
	for _, v := range conditions {
		switch v.Type {
		case mock.StepTypeInput:
			if _, ok := mock.StepInputFromString(v.Value); !ok {
				return fmt.Errorf("undefined input type %q", v.Value)
			}
		case mock.StepTypeFormat:
			if _, ok := mock.StepFormatFromString(v.Value); !ok {
				return fmt.Errorf("undefined input format type %q", v.Value)
			}
		case mock.StepTypeArray:
			value, err := strconv.ParseFloat(v.Value, 64)
			if err != nil || value < 1 {
				return fmt.Errorf("non valid index value %q", v.Value)
			}
		case mock.StepTypeOperator:
			if _, ok := mock.StepOperatorFromString(v.Value); !ok {
				return fmt.Errorf("undefined operator %q", v.Value)
			}
		}
	}

	return nil
}
