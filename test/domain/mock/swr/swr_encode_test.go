package swr_test

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	"github.com/Rafael24595/go-api-core/test/support/assert"
)

func TestMarshal_Simple(t *testing.T) {
	conds := []string{
		"input",
		"payload.json.lang",
		"payload.json.items.[1].$eq.<zig>",
		"payload.json.value.$eq.<>",
		`payload.json.rate.$eq.<0\.3>`,
	}

	for _, cond := range conds {
		t.Run(cond, func(t *testing.T) {
			steps, errs := swr.Unmarshal(cond)
			assert.Len(t, 0, errs)

			result, errs := swr.Marshal(steps)
			assert.Len(t, 0, errs)

			assert.Equal(t, cond, result)
		})
	}
}

func TestMarshal_InvalidInput(t *testing.T) {
	cond := "input"

	steps, errs := swr.Unmarshal(cond)
	assert.Len(t, 0, errs)

	_, errs = swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined input type "input"`, errs[0].Error())
}

func TestMarshal_InputMissingFormat(t *testing.T) {
	cond := "payload.lang"

	opts := swr.MarshalOpts{
		Evalue: true,
	}

	steps, errs := swr.Unmarshal(cond)
	assert.Len(t, 0, errs)

	_, errs = swr.MarshalWithOptions(steps, opts)

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined input format type "lang"`, errs[0].Error())

	cond = "arguments.lang"
	steps, errs = swr.Unmarshal(cond)
	assert.Len(t, 0, errs)

	_, errs = swr.MarshalWithOptions(steps, opts)

	assert.Len(t, 0, errs)
}

func TestMarshal_InvalidFormat(t *testing.T) {
	cond := "payload.csvt"

	steps, errs := swr.Unmarshal(cond)
	assert.Len(t, 0, errs)

	_, errs = swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined input format type "csvt"`, errs[0].Error())
}

func TestMarshal_InvalidOperator(t *testing.T) {
	cond := "payload.xml.lang.$rs.<rust>"

	steps, errs := swr.Unmarshal(cond)
	assert.Len(t, 0, errs)

	_, errs = swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `undefined operator "rs"`, errs[0].Error())
}

func TestMarshal_InvalidPosition(t *testing.T) {
	cond := "payload.vec_json.[-1].$eq.<golang>"

	steps, errs := swr.Unmarshal(cond)
	assert.Len(t, 0, errs)

	_, errs = swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `non valid index value "-1"`, errs[0].Error())
}

func TestMarshal_InvalidArrayIndexNonNumeric(t *testing.T) {
	cond := "payload.json.items.[abc].$eq.<zig>"

	steps, errs := swr.Unmarshal(cond)
	assert.Len(t, 0, errs)

	_, errs = swr.MarshalWithOptions(steps, swr.MarshalOpts{
		Evalue: true,
	})

	assert.Len(t, 1, errs)
	assert.Equal(t, `non valid index value "abc"`, errs[0].Error())
}
