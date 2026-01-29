package swr

import (
	"errors"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

type UnmarshalOpts struct {
	Evalue bool
}

func DefaultUnmarshalOpts() UnmarshalOpts {
	return UnmarshalOpts{
		Evalue: false,
	}
}

func Unmarshal(cond string) ([]Step, []error) {
	return unmarshal(cond, DefaultUnmarshalOpts())
}

func UnmarshalWithOptions(cond string, opts UnmarshalOpts) ([]Step, []error) {
	return unmarshal(cond, opts)
}

func unmarshal(cond string, opts UnmarshalOpts) ([]Step, []error) {
	steps := make([]Step, 0)
	cond = strings.TrimSpace(cond)

	if cond == "" {
		return steps, make([]error, 0)
	}

	fragments := collection.VectorFromList(utils.SplitByRune(cond, '.'))

	headers, err := findHeaders(fragments)
	if err != nil {
		return steps, []error{err}
	}

	steps = append(steps, headers...)

	var prevStep *Step
	for fragments.Size() > 0 {
		cursor, ok := fragments.Shift()
		if !ok {
			break
		}

		var step *Step
		if _, value, ok := strings.Cut(cursor, "$"); ok {
			step = NewConditionStep(StepTypeOperator, value)
		} else if value, ok := findPosition(cursor); ok {
			step = NewConditionStep(StepTypeArray, value)
		} else if value, ok := findValue(cursor); ok {
			step = NewConditionStep(StepTypeValue, value)
		} else if value, ok := findInput(cursor, prevStep); ok {
			step = NewConditionStep(StepTypeInput, value)
		} else if value, ok := findFormat(cursor, prevStep); ok {
			step = NewConditionStep(StepTypeFormat, value)
		} else {
			step = NewConditionStep(StepTypeField, cursor)
		}

		steps = append(steps, *step)

		prevStep = step
	}

	steps = FixStepsOrder(steps)

	if opts.Evalue {
		if errs := evalueSteps(steps); len(errs) > 0 {
			return steps, errs
		}
	}

	return steps, nil
}

func findHeaders(fragments *collection.Vector[string]) ([]Step, error) {
	steps := make([]Step, 0)
	target, ok := fragments.Shift()
	if !ok {
		return steps, nil
	}

	step := NewConditionStep(StepTypeInput, target)
	steps = append(steps, *step)

	if target != string(StepInputPayload) {
		return steps, nil
	}

	format, ok := fragments.Shift()
	if !ok {
		return steps, errors.New("payload format is undefined")
	}

	step = NewConditionStep(StepTypeFormat, format)
	steps = append(steps, *step)

	return steps, nil
}
