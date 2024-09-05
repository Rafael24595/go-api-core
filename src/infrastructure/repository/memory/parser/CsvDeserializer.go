package parser

import "strconv"

type CsvDeserializer struct {
	Structures map[string][]string
}

func (s *CsvDeserializer) Deserialize(csv string) {
	tables := readTables(csv)
	println(tables)
}

func readTables(csv string) []ResourceNexus {
	raw := csv
	tables := []ResourceNexus{}
	for len(raw) > 0 {
		var table string
		table, raw = readNextTable(raw)
		if len(table) > 0 {
			nexus := parseTable(table)
			tables = append(tables, nexus)
		}
	}
	return tables
}

func readNextTable(csv string) (string, string) {
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

func parseTable(table string) ResourceNexus {
	rootCount := 0
	headCount := 0
	parseHead := false

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
			if len(bufferBase) == 0 && isArrowComponent(v) {
				//
			} else if v == '\n' {
				if row == 0 {
					heads = parseHeaders(bufferBase)
				} else {
					row := parseRow(bufferBase, len(heads) != 0)
					items = append(items, row)
				}
				bufferBase = ""
				row++
			} else {
				bufferBase += string(v)
			}
		}
	}

	return ResourceNexus{
		key: name,
		headers: heads,
		nodes: items,
	}
}

func parseHeaders(row string) []string {
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
	return heads
}

func parseRow(row string, header bool) ResourceGroup {
	instance := instanceOf(row, header)
	switch instanceOf(row, header) {
	case "MAP":
		return ResourceGroup{category: instance, group: parseMap(row)}
	case "ARR":
		return ResourceGroup{category: instance, group: parseArr(row)}
	case "STR":
		parseStr(row)
	case "ENU":
		parseEnu(row)
	}
	return ResourceGroup{}
}

func instanceOf(row string, header bool) string {
	inString := false
	escape := false
	for _, v := range row {
		if !inString {
			if v == '"' {
				inString = !inString
			} else if v == MAP_LINKER {
				return "MAP"
			} else if v == ARR_SEPARATOR || v == ARR_CLOSING {
				return "ARR"
			} else if v == STR_SEPARATOR || v == STR_CLOSING {
				return "STR"
			}
		} else if !escape {
			if v == '"' {
				inString = !inString
			} else if v == '\\' {
				escape = true
			}
		} else {
			escape = false
		}
	}
	if !header {
		return "ENU"
	}
	return "STR"
}

func parseMap(row string) map[string]ResourceNode {
	mapp := map[string]ResourceNode{}
	key := ""
	buffer := ""
	for _, v := range row {
		if v == MAP_LINKER {
			node := parseObj(buffer)
			key = node.key()
			buffer = ""
		} else if v == MAP_SEPARATOR {
			node := parseObj(buffer)
			mapp[key] = node
			buffer = ""
		} else {
			buffer += string(v)
		}
	}
	node := parseObj(buffer)
	mapp[key] = node
	return mapp
}

func parseArr(row string) []ResourceNode {
	arr := []ResourceNode{}
	buffer := ""
	for _, v := range row {
		if v == ARR_SEPARATOR {
			node := parseObj(buffer)
			arr = append(arr, node)
			buffer = ""
		} else if v == ARR_CLOSING {
			break
		} else {
			buffer += string(v)
		}

	}
	node := parseObj(buffer)
	arr = append(arr, node)
	return arr
}

func parseStr(row string) {

}

func parseEnu(row string) {

}

func parseObj(obj string) ResourceNode {
	len := len(obj)
	if len == 0 {
		return fromNonPointer("")
	}
	if obj[0] == byte(PTR_HEADER) {
		return parsePtr(obj)
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

func parsePtr(obj string) ResourceNode {
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
	index, err :=strconv.Atoi(buffer)
	if err != nil {
		panic(err)
	}
	return fromPointer(key, index)
}

func isArrowComponent(v rune) bool {
	return (v >= '0' && v <= '9') || v == 'H' || v == '-' || v == '>' || v == ' ' || v == '\t'
}