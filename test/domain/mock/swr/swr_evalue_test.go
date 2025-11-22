package swr_test

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	"github.com/Rafael24595/go-api-core/test/support/assert"
)

func TestEvalueStepPosition_ValidFlow(t *testing.T) {
	cond := "payload.json.items.[1].$eq.<zig>"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 0, errs)
}

func TestEvalueStepPosition_FirstElementNotInput(t *testing.T) {
	steps := []swr.Step{
		{
			Order: 0,
			Type:  swr.StepTypeOperator,
			Value: "eq",
		},
	}

	_, errs := swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `first element should be input type`, errs[0].Error())
}

func TestEvalueStepPosition_InputInMiddle(t *testing.T) {
	steps := []swr.Step{
		{
			Order: 0,
			Type:  swr.StepTypeInput,
			Value: string(swr.StepInputPayload),
		},
		{
			Order: 1,
			Type:  swr.StepTypeFormat,
			Value: "json",
		},
		{
			Order: 2,
			Type:  swr.StepTypeField,
			Value: "1",
		},
		{
			Order: 3,
			Type:  swr.StepTypeInput,
			Value: string(swr.StepInputPayload),
		},
	}

	_, errs := swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `an input operation cannot be applied in the middle of an operation, but input found on 3 position`, errs[0].Error())
}

func TestEvalueStepPosition_DoubleOperator(t *testing.T) {
	cond := "payload.json.lang.$ne.$eq.<zig>"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `a compare operation is required after operator, but operator found on 4 position`, errs[0].Error())
}

func TestEvalueStepPosition_CompareWithoutLeftOperand(t *testing.T) {
	cond := "payload.json.$eq.<golang>"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `a comparable value is required before compare operator, but format found on 1 position`, errs[0].Error())
}

func TestEvalueStepPosition_CompareWithoutRightOperand(t *testing.T) {
	cond := "payload.json.lang.$eq.[0]"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `a comparable value is required after compare operator, but array found on 4 position`, errs[0].Error())
}

func TestEvalueStepPosition_ValueExtraction(t *testing.T) {
	cond := `payload.json.<0\.11>.lang`

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 2, errs)
	assert.Equal(t, `a defined value cannot be extracted from a structure type, but format found on 1 position`, errs[0].Error())
	assert.Equal(t, `a value cannot be extracted from a flat value, but value found on 2 position`, errs[1].Error())
}

func TestEvalueStepPosition_FormatArgument(t *testing.T) {
	steps := []swr.Step{
		{
			Order: 0,
			Type:  swr.StepTypeInput,
			Value: string(swr.StepInputArgument),
		},
		{
			Order: 1,
			Type:  swr.StepTypeFormat,
			Value: "json",
		},
		{
			Order: 2,
			Type:  swr.StepTypeField,
			Value: "1",
		},
		{
			Order: 3,
			Type:  swr.StepTypeOperator,
			Value: string(swr.StepOperatorEq),
		},
		{
			Order: 4,
			Type:  swr.StepTypeValue,
			Value: "zig",
		},
	}

	_, errs := swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `an header input cannot be formatted, but format found on 1 position`, errs[0].Error())
}
