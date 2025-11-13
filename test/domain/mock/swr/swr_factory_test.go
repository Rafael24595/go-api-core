package swr_test

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	"github.com/Rafael24595/go-api-core/test/support/assert"
)

func TestUnmarshal_SimpleInput(t *testing.T) {
	cond := "input"
	steps, err := swr.Unmarshal(cond)

	assert.NoError(t, err)
	assert.Len(t, 1, steps)
	assert.Equal(t, mock.StepTypeInput, steps[0].Type)
	assert.Equal(t, "input", steps[0].Value)
}

func TestUnmarshal_PayloadWithFormat(t *testing.T) {
	cond := "payload.json.lang"
	steps, err := swr.Unmarshal(cond)

	assert.NoError(t, err)
	assert.Len(t, 3, steps)

	assert.Equal(t, mock.StepTypeInput, steps[0].Type)
	assert.Equal(t, "payload", steps[0].Value)

	assert.Equal(t, mock.StepTypeFormat, steps[1].Type)
	assert.Equal(t, "json", steps[1].Value)

	assert.Equal(t, mock.StepTypeField, steps[2].Type)
	assert.Equal(t, "lang", steps[2].Value)
}

func TestUnmarshal_OperatorArrayValue(t *testing.T) {
	cond := "payload.json.items.[1].$eq.<zig>"
	steps, err := swr.Unmarshal(cond)

	assert.NoError(t, err)
	assert.Len(t, 6, steps)

	assert.Equal(t, mock.StepTypeArray, steps[3].Type)
	assert.Equal(t, "1", steps[3].Value)

	assert.Equal(t, mock.StepTypeOperator, steps[4].Type)
	assert.Equal(t, "eq", steps[4].Value)

	assert.Equal(t, mock.StepTypeValue, steps[5].Type)
	assert.Equal(t, "zig", steps[5].Value)
}

func TestUnmarshal_InvalidInput(t *testing.T) {
	cond := "input"

	_, err := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Error(t, err)
	assert.Equal(t, `undefined input type "input"`, err.Error())
}

func TestUnmarshal_InputMissingFormat(t *testing.T) {
	cond := "payload.lang"

	opts := swr.UnmarshalOpts{
		Evalue: true,
	}

	_, err := swr.UnmarshalWithOptions(cond, opts)

	assert.Error(t, err)
	assert.Equal(t, `undefined input format type "lang"`, err.Error())

	cond = "arguments.lang"
	_, err = swr.UnmarshalWithOptions(cond, opts)

	assert.NoError(t, err)
}

func TestUnmarshal_InvalidFormat(t *testing.T) {
	cond := "payload.csvt"

	_, err := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Error(t, err)
	assert.Equal(t, `undefined input format type "csvt"`, err.Error())
}

func TestUnmarshal_InvalidOperator(t *testing.T) {
	cond := "payload.xml.lang.$rs.<rust>"

	_, err := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Error(t, err)
	assert.Equal(t, `undefined operator "rs"`, err.Error())
}

func TestUnmarshal_InvalidPosition(t *testing.T) {
	cond := "payload.vec_json.[-1].$eq.<golang>"

	_, err := swr.UnmarshalWithOptions(cond, swr.UnmarshalOpts{
		Evalue: true,
	})

	assert.Error(t, err)
	assert.Equal(t, `non valid index value "-1"`, err.Error())
}
