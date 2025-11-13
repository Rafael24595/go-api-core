package swr

import (
	"strings"
)

const (
	// REGEX_VECTOR_INDEX matches array/vector index expressions like [0], [1], etc.
	REGEX_VECTOR_INDEX = `^\[(.*)\]$`
	// REGEX_RAW_VALUE matches raw literal values wrapped in angle brackets like <value>.
	REGEX_RAW_VALUE = `^<(.*)>$`
)

type StepType string

const (
	StepTypeInput    StepType = "input"
	StepTypeFormat   StepType = "format"
	StepTypeArray    StepType = "array"
	StepTypeField    StepType = "field"
	StepTypeValue    StepType = "value"
	StepTypeOperator StepType = "operator"
)

type StepInput string

const (
	StepInputPayload   StepInput = "payload"
	StepInputArguments StepInput = "arguments"
)

func StepInputFromString(s string) (StepInput, bool) {
	switch s {
	case string(StepInputPayload):
		return StepInputPayload, true
	case string(StepInputArguments):
		return StepInputArguments, true
	default:
		return "", false
	}
}

type StepFormat string

const (
	StepFormatText    StepFormat = "text"
	StepFormatJson    StepFormat = "json"
	StepFormatVecJson StepFormat = "vec_json"
	StepFormatXml     StepFormat = "xml"
	StepFormatVecXml  StepFormat = "vec_xml"
)

func StepFormatFromString(s string) (StepFormat, bool) {
	switch s {
	case string(StepFormatText):
		return StepFormatText, true
	case string(StepFormatJson):
		return StepFormatJson, true
	case string(StepFormatVecJson):
		return StepFormatVecJson, true
	case string(StepFormatXml):
		return StepFormatXml, true
	case string(StepFormatVecXml):
		return StepFormatVecXml, true
	default:
		return "", false
	}
}

type StepOperator string

const (
	// StepOperatorEq represents the equality operator "$eq".
	StepOperatorEq StepOperator = "$eq"
	// StepOperatorNeE represents the inequality operator "$ne".
	StepOperatorNe StepOperator = "$ne"
	// StepOperatorGt represents the greater-than operator "$gt".
	StepOperatorGt StepOperator = "$gt"
	// StepOperatorGte represents the greater-than-or-equal operator "$gte".
	StepOperatorGte StepOperator = "$gte"
	// StepOperatorLt represents the less-than operator "$lt".
	StepOperatorLt StepOperator = "$lt"
	// StepOperatorLte represents the less-than-or-equal operator "$lte".
	StepOperatorLte StepOperator = "$lte"
	// StepOperatorAnd represents the logical AND operator "$and".
	StepOperatorAnd StepOperator = "$and"
	// StepOperatorOr represents the logical OR operator "$or".
	StepOperatorOr StepOperator = "$or"
)

func StepOperatorFromString(s string) (StepOperator, bool) {
	switch s {
	case string(StepOperatorEq):
		return StepOperatorEq, true
	case string(StepOperatorNe):
		return StepOperatorNe, true
	case string(StepOperatorGt):
		return StepOperatorGt, true
	case string(StepOperatorGte):
		return StepOperatorGte, true
	case string(StepOperatorLt):
		return StepOperatorLt, true
	case string(StepOperatorLte):
		return StepOperatorLte, true
	case string(StepOperatorAnd):
		return StepOperatorAnd, true
	case string(StepOperatorOr):
		return StepOperatorOr, true
	default:
		return "", false
	}
}

func StepOperatorFromUnsignedString(s string) (StepOperator, bool) {
	if !strings.HasPrefix(s, "$") {
		s = "$" + s
	}
	return StepOperatorFromString(s)
}
