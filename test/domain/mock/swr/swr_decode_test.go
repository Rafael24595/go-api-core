package swr_test

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	"github.com/Rafael24595/go-api-core/test/support/assert"
)

func TestUnmarshal_EmptyString(t *testing.T) {
	steps, errs := swr.Unmarshal("")

	assert.Len(t, 0, errs)
	assert.Len(t, 0, steps)
}

func TestUnmarshal_SimpleInput(t *testing.T) {
	cond := "input"
	steps, errs := swr.Unmarshal(cond)

	assert.Len(t, 0, errs)
	assert.Len(t, 1, steps)

	assert.Equal(t, swr.StepTypeInput, steps[0].Type)
	assert.Equal(t, "input", steps[0].Value)
}

func TestUnmarshal_PayloadWithFormat(t *testing.T) {
	cond := "payload.json.lang"
	steps, errs := swr.Unmarshal(cond)

	assert.Len(t, 0, errs)
	assert.Len(t, 3, steps)

	assert.Equal(t, swr.StepTypeInput, steps[0].Type)
	assert.Equal(t, "payload", steps[0].Value)

	assert.Equal(t, swr.StepTypeFormat, steps[1].Type)
	assert.Equal(t, "json", steps[1].Value)

	assert.Equal(t, swr.StepTypeField, steps[2].Type)
	assert.Equal(t, "lang", steps[2].Value)
}

func TestUnmarshal_OperatorArrayValue(t *testing.T) {
	cond := "payload.json.items.[1].$eq.<zig>"
	steps, errs := swr.Unmarshal(cond)

	assert.Len(t, 0, errs)
	assert.Len(t, 6, steps)

	assert.Equal(t, swr.StepTypeArray, steps[3].Type)
	assert.Equal(t, "1", steps[3].Value)

	assert.Equal(t, swr.StepTypeOperator, steps[4].Type)
	assert.Equal(t, "eq", steps[4].Value)

	assert.Equal(t, swr.StepTypeValue, steps[5].Type)
	assert.Equal(t, "zig", steps[5].Value)
}

func TestUnmarshal_ArgumentsFlow(t *testing.T) {
	cond := "arguments.lang"
	steps, errs := swr.Unmarshal(cond)

	assert.Len(t, 0, errs)
	assert.Len(t, 2, steps)

	assert.Equal(t, swr.StepTypeInput, steps[0].Type)
	assert.Equal(t, "arguments", steps[0].Value)

	assert.Equal(t, swr.StepTypeField, steps[1].Type)
	assert.Equal(t, "lang", steps[1].Value)
}

func TestUnmarshal_EmptyValue(t *testing.T) {
	cond := "payload.json.value.$eq.<>"

	steps, errs := swr.Unmarshal(cond)

	assert.Len(t, 0, errs)
	assert.Len(t, 5, steps)

	assert.Equal(t, swr.StepTypeValue, steps[4].Type)
	assert.Equal(t, "", steps[4].Value)
}

func TestUnmarshal_InvalidInput(t *testing.T) {
	cond := "input"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined input type "input"`, errs[0].Error())
}

func TestUnmarshal_InputMissingFormat(t *testing.T) {
	cond := "payload.lang"

	opts := swr.UnmarshalOpts{
		Evalue: true,
	}

	_, errs := swr.UnmarshalWithOptions(cond, opts)

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined input format type "lang"`, errs[0].Error())

	cond = "arguments.lang"
	_, errs = swr.UnmarshalWithOptions(cond, opts)

	assert.Len(t, 1, errs)
}

func TestUnmarshal_InvalidFormat(t *testing.T) {
	cond := "payload.csvt"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined input format type "csvt"`, errs[0].Error())
}

func TestUnmarshal_InvalidOperator(t *testing.T) {
	cond := "payload.xml.lang.$rs.<rust>"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined operator "rs"`, errs[0].Error())
}

func TestUnmarshal_InvalidPosition(t *testing.T) {
	cond := "payload.vec_json.[-1].$eq.<golang>"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `non valid index value "-1"`, errs[0].Error())
}

func TestUnmarshal_InvalidArrayIndexNonNumeric(t *testing.T) {
	cond := "payload.json.items.[abc].$eq.<zig>"

	_, errs := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `non valid index value "abc"`, errs[0].Error())
}

