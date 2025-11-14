package swr

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

func findValue(cursor string) (string, bool) {
	re := regexp.MustCompile(REGEX_RAW_VALUE)
	matches := re.FindStringSubmatch(cursor)
	if len(matches) < 2 {
		return "", false
	}
	return matches[1], true
}

func findPosition(cursor string) (string, bool) {
	re := regexp.MustCompile(REGEX_VECTOR_INDEX)
	matches := re.FindStringSubmatch(cursor)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

func evalueSteps(steps []Step) []error {
	errors := make([]error, 0)

	var prev *Step
	for _, s := range steps {
		if err := evalueStepType(s); err != nil {
			errors = append(errors, err)
		}

		if err := evalueStepPosition(s, prev); err != nil {
			errors = append(errors, err)
		}

		prev = &s
	}

	return errors
}

func evalueStepType(step Step) error {
	switch step.Type {
	case StepTypeInput:
		if _, ok := StepInputFromString(step.Value); !ok {
			return fmt.Errorf("undefined input type %q", step.Value)
		}
	case StepTypeFormat:
		if _, ok := StepFormatFromString(step.Value); !ok {
			return fmt.Errorf("undefined input format type %q", step.Value)
		}
	case StepTypeArray:
		value, err := strconv.ParseFloat(step.Value, 64)
		if err != nil || value < 0 {
			return fmt.Errorf("non valid index value %q", step.Value)
		}
	case StepTypeOperator:
		if _, ok := StepOperatorFromUnsignedString(step.Value); !ok {
			return fmt.Errorf("undefined operator %q", step.Value)
		}
	}

	return nil
}

func evalueStepPosition(cursor Step, parent *Step) error {
	if parent == nil {
		if cursor.Type != StepTypeInput {
			return errors.New("first element should be input type")
		}
		return nil
	}

	if parent.Type != StepTypeOperator && cursor.Type == StepTypeInput {
		return fmt.Errorf(`an input operation cannot be applied in the middle of an operation, but %s found on %d position`, cursor.Type, cursor.Order)
	}

	if parent.Type == StepTypeOperator && cursor.Type == StepTypeOperator {
		return fmt.Errorf(`a compare operation is required after operator, but %s found on %d position`, cursor.Type, cursor.Order)
	}

	if isFormatedInput(*parent) && cursor.Type != StepTypeFormat {
		return fmt.Errorf(`a formatted input requires a format specification, but %s found on %d position`, cursor.Type, cursor.Order)
	}

	if isLogicalOperator(*parent) && !isComparableRight(cursor) {
		return fmt.Errorf(`a compare operation is required after logical operator, but %s found on %d position`, cursor.Type, cursor.Order)
	}

	if isCompareOperator(*parent) && !isComparableRight(cursor) {
		return fmt.Errorf(`a comparable value is required after compare operator, but %s found on %d position`, cursor.Type, cursor.Order)
	}

	if isCompareOperator(cursor) && !isComparableLeft(*parent) {
		return fmt.Errorf(`a comparable value is required before compare operator, but %s found on %d position`, parent.Type, parent.Order)
	}

	if parent.Type == StepTypeValue && cursor.Type != StepTypeOperator {
		return fmt.Errorf(`a value cannot be extracted from a flat value, but %s found on %d position`, parent.Type, parent.Order)
	}

	if cursor.Type == StepTypeValue && parent.Type != StepTypeOperator {
		return fmt.Errorf(`a defined value cannot be extracted from a structure type, but %s found on %d position`, parent.Type, parent.Order)
	}

	return nil
}

func isFormatedInput(step Step) bool {
	if step.Type != StepTypeInput {
		return false
	}

	switch step.Value {
	case string(StepInputPayload):
		return true
	default:
		return false
	}
}

func isLogicalOperator(step Step) bool {
	if step.Type != StepTypeOperator {
		return false
	}

	operator, ok := StepOperatorFromUnsignedString(step.Value)
	if !ok {
		return false
	}

	switch operator {
	case StepOperatorAnd,
		StepOperatorOr:
		return true
	}
	return false
}

func isCompareOperator(step Step) bool {
	if step.Type != StepTypeOperator {
		return false
	}

	operator, ok := StepOperatorFromUnsignedString(step.Value)
	if !ok {
		return false
	}

	switch operator {
	case StepOperatorEq,
		StepOperatorNe,
		StepOperatorGt,
		StepOperatorGte,
		StepOperatorLt,
		StepOperatorLte:
		return true
	}
	return false
}

func isComparableLeft(step Step) bool {
	switch step.Type {
	case StepTypeArray,
		StepTypeField,
		StepTypeValue:
		return true
	}
	return false
}

func isComparableRight(step Step) bool {
	switch step.Type {
	case StepTypeInput,
		StepTypeValue:
		return true
	}
	return false
}
