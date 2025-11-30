package swr

import (
	"strings"
)

type MarshalOpts struct {
	Evalue bool
}

func DefaultMarshalOpts() MarshalOpts {
	return MarshalOpts{
		Evalue: false,
	}
}

func Marshal(steps []Step) (string, []error) {
	return marshal(steps, DefaultMarshalOpts())
}

func MarshalWithOptions(steps []Step, opts MarshalOpts) (string, []error) {
	return marshal(steps, opts)
}

func marshal(steps []Step, opts MarshalOpts) (string, []error) {
	ordered := OrderSteps(steps)

	if opts.Evalue {
		ordered = FixStepsOrder(ordered)
		if errs := evalueSteps(ordered); len(errs) > 0 {
			return "", errs
		}
	}

	fragments := make([]string, len(ordered))

	for i := range ordered {
		fragments[i] = toString(ordered[i])
	}

	return strings.Join(fragments, "."), nil
}

func toString(step Step) string {
	value := strings.ReplaceAll(step.Value, ".", "\\.")

	switch step.Type {
	case StepTypeArray:
		return "[" + value + "]"
	case StepTypeValue:
		return "<" + value + ">"
	case StepTypeOperator:
		return "$" + value
	}

	return value
}
