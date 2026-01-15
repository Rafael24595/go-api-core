package utils

import (
	"strings"
	"testing"

	"github.com/Rafael24595/go-api-core/test/support/assert"
)

func TestNewTable(t *testing.T) {
	table := NewTable()

	assert.Equal(t, 0, len(table.headers))
	assert.Equal(t, 0, len(table.cols))
}

func TestHeaders_AddsUniqueHeaders(t *testing.T) {
	table := NewTable()

	table.Headers("ID", "Name", "ID")

	assert.Equal(t, 2, len(table.headers))

	assert.Equal(t,
		strings.Join([]string{"ID", "Name"}, " "),
		strings.Join(table.headers, " "))
}

func TestField_SetsDataCorrectly(t *testing.T) {
	table := NewTable()

	table.Headers("Id", "Lang")

	table.Field("Id", 0, 1).
		Field("Lang", 0, "Golang")

	assert.Equal(t, 2, len(table.cols))

	assert.Equal(t, 1, len(table.cols["Id"]))
	assert.Equal(t, 1, len(table.cols["Lang"]))

	assert.Equal(t, table.cols["Id"][0], "1")
	assert.Equal(t, table.cols["Lang"][0], "Golang")
}

func TestField_IgnoresUnknownHeader(t *testing.T) {
	table := NewTable()

	table.Headers("Id")

	assert.Equal(t, len(table.headers), 1)

	table.Field("Unknown", 0, "test")

	assert.Equal(t, len(table.headers), 1)
	assert.Equal(t, len(table.cols["Id"]), 0)
}

func TestCalcSize(t *testing.T) {
	table := NewTable()

	table.Headers("Id", "Lang")

	table.Field("Id", 0, 1).
		Field("Lang", 0, "Clojure")

	size := table.calcSize()

	assert.Equal(t, len("Id"), size["Id"])
	assert.Equal(t, len("Clojure"), size["Lang"])
}

func TestToString_SimpleTable(t *testing.T) {
	table := NewTable()

	table.Headers("Id", "Lang")

	table.Field("Id", 0, 1).
		Field("Lang", 0, "Zig")

	table.Field("Id", 1, 2).
		Field("Lang", 1, "Golang")

	expected := "" +
		"| Id |  Lang  |\n" +
		"|----|--------|\n" +
		"| 1  | Zig    |\n" +
		"| 2  | Golang |"

	assert.Equal(t, expected, table.ToString())
}

func TestToString_IncompleteTable(t *testing.T) {
	table := NewTable()

	table.Headers("Id", "Lang")

	table.Field("Lang", 0, "Zig")

	table.Field("Id", 1, 2)

	expected := "" +
		"| Id | Lang |\n" +
		"|----|------|\n" +
		"|    | Zig  |\n" +
		"| 2  |      |"

	assert.Equal(t, expected, table.ToString())
}

func TestToString_EmptyRows(t *testing.T) {
	table := NewTable()

	table.Headers("Id", "Lang")

	expected :=
		"| Id | Lang |\n" +
		"|----|------|"

	assert.Equal(t, expected, table.ToString())
}
