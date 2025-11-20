package response_test

import (
	"fmt"
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-api-core/test/support/assert"
)

func TestFixResponses_AddDefault(t *testing.T) {
	input := []mock.Response{
		{Order: 0, Name: "test"},
	}

	result := mock.FixResponses(input)

	assert.Len(t, 2, result)

	assert.Equal(t, mock.DefaultResponse, result[0].Name)
}

func TestFixResponses_IgnoreDuplicateDefault(t *testing.T) {
	input := []mock.Response{
		{Order: 0, Name: "test-1"},
		{Order: 1, Name: "test-2"},
		{Order: 2, Name: "test-3"},
		{Order: 3, Name: "test-4"},
		{Order: 4, Name: mock.DefaultResponse},
		{Order: 8, Name: mock.DefaultResponse},
		{Order: 7, Name: mock.DefaultResponse},
		{Order: 6, Name: mock.DefaultResponse},
		{Order: 5, Name: "test-5"},
	}

	result := mock.FixResponses(input)

	assert.Len(t, 6, result)

	for i := 1; i < len(result); i++ {
		assert.Equal(t, fmt.Sprintf("test-%d", i), result[i].Name)
		assert.Equal(t, i, result[i].Order)
	}
}

func TestFixResponses_OrderDefault(t *testing.T) {
	input := []mock.Response{
		{Order: 0, Name: "test-0"},
		{Order: 1, Name: "test-1"},
		{Order: 2, Name: "test-2"},
		{Order: 3, Name: "test-3"},
		{Order: 4, Name: mock.DefaultResponse, Status: false},
		{Order: 5, Name: "test-4"},
	}

	result := mock.FixResponses(input)

	assert.Len(t, 6, result)

	assert.Equal(t, mock.DefaultResponse, result[0].Name)
	assert.Equal(t, true, result[0].Status)
}

func TestFixResponses_Order(t *testing.T) {
	input := []mock.Response{
		{Order: 5, Name: "test-5"},
		{Order: 2, Name: "test-2"},
		{Order: 4, Name: "test-4"},
		{Order: 3, Name: "test-3"},
		{Order: 1, Name: "test-1"},
	}

	result := mock.FixResponses(input)

	assert.Len(t, 6, result)

	for i := 1; i < len(result); i++ {
		assert.Equal(t, fmt.Sprintf("test-%d", i), result[i].Name)
	}
}

func TestFixResponses_FitOrder(t *testing.T) {
	input := []mock.Response{
		{Order: 2, Name: "test-2"},
		{Order: 1, Name: "test-1"},
		{Order: 3, Name: "test-3"},
		{Order: 5, Name: "test-4"},
	}

	result := mock.FixResponses(input)

	assert.Len(t, 5, result)

	for i := 1; i < len(result); i++ {
		assert.Equal(t, fmt.Sprintf("test-%d", i), result[i].Name)
		assert.Equal(t, i, result[i].Order)
	}
}

func TestFixResponsesName_Unique(t *testing.T) {
	name := "test"

	input := []mock.Response{
		{Order: 0, Name: name},
		{Order: 1, Name: name},
		{Order: 2, Name: name},
	}

	expected := []string{
		"default", "test", "test-copy-1", "test-copy-2",
	}

	result := mock.FixResponses(input)

	for i := 0; i < len(result); i++ {
		assert.Equal(t, expected[i], result[i].Name)
	}
}

func TestFixResponsesName_ExistingCopy(t *testing.T) {
	name := "test"

	input := []mock.Response{
		{Order: 0, Name: name},
		{Order: 1, Name: name},
		{Order: 2, Name: "test-copy-1"},
		{Order: 2, Name: name},
	}

	expected := []string{
		"default", "test", "test-copy-1", "test-copy-1-copy-1", "test-copy-2",
	}

	result := mock.FixResponses(input)

	for i := 0; i < len(result); i++ {
		assert.Equal(t, expected[i], result[i].Name)
	}
}
