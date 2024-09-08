package parser

import (
	"strconv"
	"strings"
)

type csvReader struct {
	structures map[string][]string
}

func newDeserializerReader() *csvReader {
	return &csvReader{
		structures: make(map[string][]string),
	}
}

func (r *csvReader) read(csv string) ResourceCollection  {
	raw := csv
	tables := map[string]ResourceNexus{}
	for len(raw) > 0 {
		var table string
		table, raw = r.readNext(raw)
		if len(strings.TrimSpace(table)) > 0 {
			nexus := r.parse(table)
			tables[nexus.key] = nexus
		}
	}
	return newCollection(tables)
}

func (d *csvReader) readNext(csv string) (string, string) {
	initial := 0
	headCount := 0

	end := 0
	tailCount := 0

	count := 0
	for i, v := range csv {
		count = i
		if v == '/' {
			if headCount == 0 {
				initial = i
			}
			headCount++
		} else if v == '*' && headCount > 0 {
			headCount++
		} else if v == '\n' {
			if tailCount == 1 {
				end = i
				break
			}
			tailCount++
		} else {
			if headCount < 3 {
				headCount = 0
			}
			if tailCount < 2 {
				tailCount = 0
			}
		}
	}

	if count == len(csv)-1 {
		end = count
	}

	if end == 0 {
		return "", ""
	}

	return csv[initial:end], csv[end:]
}

func (d *csvReader) parse(table string) ResourceNexus {
	rootCount := 0
	headCount := 0
	parseHead := false

	arrowCount := 0
	passArrow := false

	name := ""
	heads := []string{}
	items := []ResourceGroup{}

	bufferBase := ""
	row := 0
	for _, v := range table {
		if !parseHead {
			if v == '/' {
				headCount++
			} else if v == '*' {
				rootCount++
				headCount++
			} else if v == '\n' && headCount == 3 {
				parseHead = true
				name = bufferBase
				bufferBase = ""
			} else if v == ' ' {
				//
			} else {
				bufferBase += string(v)
			}
		} else {
			isArrow, isHead := d.isArrowComponent(v)
			if len(bufferBase) == 0 && isArrow && !passArrow {
				if arrowCount > 0 {
					passArrow = true
				}
				if isHead {
					arrowCount++
				}
			} else if v == '\n' {
				if row == 0 {
					heads = d.parseHeaders(bufferBase)
				} else {
					row := d.parseRow(bufferBase, heads)
					items = append(items, row)
				}
				bufferBase = ""
				arrowCount = 0
				passArrow = false
				row++
			} else {
				bufferBase += string(v)
			}
		}
	}

	return newNexus(name, rootCount == 2, items)
}

func (d *csvReader) parseHeaders(row string) []string {
	heads := []string{}
	buffer := ""
	for _, v := range row {
		if v == HEA_SEPARATOR {
			heads = append(heads, buffer)
			buffer = ""
		} else {
			buffer += string(v)
		}
	}
	if len(buffer) > 0 {
		heads = append(heads, buffer)
	}
	return heads
}

func (d *csvReader) parseRow(row string, header []string) ResourceGroup {
	instance := d.instanceOf(row, len(header) != 0)
	var group interface{}
	switch instance {
	case "MAP":
		group = d.parseMap(row)
	case "ARR":
		group = d.parseArr(row)
	case "STR":
		group = d.parseStr(row)
	default:
		group = d.parseObj(row)
	}
	return newGroup(instance, header, group)
}

func (d *csvReader) instanceOf(row string, header bool) string {
	inString := false
	escape := false
	for _, v := range row {
		if !inString {
			if v == '"' {
				inString = true
			} else if v == MAP_LINKER {
				return "MAP"
			} else if v == ARR_SEPARATOR || v == ARR_CLOSING {
				return "ARR"
			} else if v == STR_SEPARATOR || v == STR_CLOSING {
				return "STR"
			}
		} else if !escape {
			if v == '"' {
				inString = false
			} else if v == '\\' {
				escape = true
			}
		} else {
			escape = false
		}
	}
	if !header {
		return "OBJ"
	}
	return "STR"
}

func (d *csvReader) parseMap(row string) map[string]ResourceNode {
	inString := false
	escape := false

	mapp := map[string]ResourceNode{}
	key := ""
	buffer := ""
	for _, v := range row {
		isSpecialRuneInString := inString && (v == MAP_LINKER || v == MAP_SEPARATOR)
		isNotSpecialRune := v != MAP_LINKER && v != MAP_SEPARATOR
		if isNotSpecialRune || isSpecialRuneInString {
			buffer += string(v)
		}
		if !inString {
			if v == '"' {
				inString = true
			} else if v == MAP_LINKER {
				node := d.parseObj(buffer)
				key = node.key()
				buffer = ""
			} else if v == MAP_SEPARATOR {
				node := d.parseObj(buffer)
				mapp[key] = node
				buffer = ""
			}
		} else if !escape {
			if v == '"' {
				inString = false
			} else if v == '\\' {
				escape = true
			}
		} else {
			escape = false
		}
	}
	node := d.parseObj(buffer)
	mapp[key] = node
	return mapp
}

func (d *csvReader) parseArr(row string) []ResourceNode {
	return d.parseLst(row, ARR_SEPARATOR, ARR_CLOSING)
}

func (d *csvReader) parseStr(row string) []ResourceNode {
	return d.parseLst(row, STR_SEPARATOR, STR_CLOSING)
}

func (d *csvReader) parseLst(row string, separator, closing rune) []ResourceNode {
	inString := false
	escape := false

	lst := []ResourceNode{}
	buffer := ""
	for _, v := range row {
		isSpecialRuneInString := inString && (v == separator || v == closing)
		isNotSpecialRune := v != separator && v != closing
		if isNotSpecialRune || isSpecialRuneInString {
			buffer += string(v)
		}
		if !inString {
			if v == '"' {
				inString = true
			} else if v == separator {
				node := d.parseObj(buffer)
				lst = append(lst, node)
				buffer = ""
			} else if v == closing {
				break
			}
		} else if !escape {
			if v == '"' {
				inString = false
			} else if v == '\\' {
				escape = true
			}
		} else {
			escape = false
		}
	}
	node := d.parseObj(buffer)
	lst = append(lst, node)
	return lst
}

func (d *csvReader) parseObj(obj string) ResourceNode {
	len := len(obj)
	if len == 0 {
		return fromNonPointer("")
	}
	if obj[0] == byte(PTR_HEADER) {
		return d.parsePtr(obj)
	}
	if obj[0] == '"' && obj[len-1] == '"' {
		return fromNonPointer(obj[1 : len-1])
	}
	rb, err := strconv.ParseBool(obj)
	if err == nil {
		return fromNonPointer(rb)
	}
	rf, err := strconv.ParseFloat(obj, 64)
	if err == nil {
		return fromNonPointer(rf)
	}
	ri, err := strconv.Atoi(obj)
	if err == nil {
		return fromNonPointer(ri)
	}
	panic("Does not recognized")
}

func (d *csvReader) parsePtr(obj string) ResourceNode {
	key := ""
	index := 0
	buffer := ""
	for i, v := range obj {
		if i == 0 {
			continue
		}
		if v == PTR_SEPARATOR {
			key = buffer
			buffer = ""
		} else {
			buffer += string(v)
		}
	}
	index, err := strconv.Atoi(buffer)
	if err != nil {
		panic(err)
	}
	return fromPointer(key, index)
}

func (d *csvReader) isArrowComponent(v rune) (bool, bool) {
	return (v >= '0' && v <= '9') || v == 'H' || v == '-' || v == '>' || v == ' ', v == '>'
}