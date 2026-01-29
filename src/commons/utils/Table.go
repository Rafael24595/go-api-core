package utils

import (
	"fmt"
	"slices"
	"strings"
)

type Table struct {
	headers []string
	cols    map[string][]string
}

func NewTable() *Table {
	return &Table{
		headers: make([]string, 0),
		cols:    make(map[string][]string),
	}
}

func (t *Table) Headers(headers ...string) *Table {
	for _, v := range headers {
		if slices.Contains(t.headers, v) {
			continue
		}

		t.headers = append(t.headers, v)
		t.cols[v] = make([]string, 0)
	}

	return t
}

func (t *Table) Field(header string, row int, data any) *Table {
	col, ok := t.cols[header]
	if !ok {
		return t
	}

	if row >= len(col) {
		for i := len(col); i <= row; i++ {
			col = append(col, "")
		}
	}

	col[row] = fmt.Sprintf("%v", data)
	t.cols[header] = col

	return t
}

func (t *Table) ToString() string {
	size := t.calcSize()

	headers := t.makeHeaders(size)
	separator := t.makeSeparator(size)
	table := t.makeTable(size)

	buffer := make([]string, 0)

	headersRow := t.formatRow(headers)

	hborder := strings.Repeat("-", len(headersRow))

	buffer = append(buffer, hborder)
	buffer = append(buffer, headersRow)
	buffer = append(buffer, t.formatRowOpts(separator, "-"))
	for _, r := range table {
		buffer = append(buffer, t.formatRow(r))
	}
	buffer = append(buffer, hborder)

	return "\n" + strings.Join(buffer, "\n") + "\n"

}

func (t *Table) formatRow(row []string) string {
	format := strings.Join(row, " | ")
	return fmt.Sprintf("| %s |", format)
}

func (t *Table) formatRowOpts(row []string, air string) string {
	separator := fmt.Sprintf("%s|%s", air, air)
	format := strings.Join(row, separator)
	return fmt.Sprintf("|%s%s%s|", air, format, air)
}

func (t *Table) makeSeparator(size map[string]int) []string {
	headers := make([]string, len(t.headers))

	for x, h := range t.headers {
		headers[x] = strings.Repeat("-", size[h])
	}

	return headers
}

func (t *Table) makeHeaders(size map[string]int) []string {
	headers := make([]string, len(t.headers))

	for x, h := range t.headers {
		headers[x] = Center(h, size[h])
	}

	return headers
}

func (t *Table) makeTable(size map[string]int) [][]string {
	colSize := 0

	for _, h := range t.headers {
		colSize = Max(colSize, len(t.cols[h]))
	}

	table := make([][]string, colSize)

	for y := range colSize {
		row := make([]string, len(t.headers))

		for x, h := range t.headers {
			size := size[h]

			col := t.cols[h]
			if y >= 0 && y < len(col) {
				row[x] = Right(col[y], size)
			} else {
				row[x] = strings.Repeat(" ", size)
			}
		}

		table[y] = row
	}

	return table
}

func (t *Table) calcSize() map[string]int {
	size := make(map[string]int)
	for _, h := range t.headers {
		if _, ok := size[h]; !ok {
			size[h] = len(h)
		}

		for _, c := range t.cols[h] {
			size[h] = Max(size[h], len(c))
		}
	}

	return size
}
