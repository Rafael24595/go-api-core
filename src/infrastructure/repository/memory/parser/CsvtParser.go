package parser

import (
	"strconv"
	"strings"
)

type csvtParser struct {
}

func newDeserializerParser() *csvtParser {
	return &csvtParser{
	}
}

func (p *csvtParser) parse(table string) ResourceNexus {
	rootCount := 0
	headCount := 0
	parseHead := false

	arrowCount := 0
	passArrow := false

	name := ""
	heads := []string{}
	items := []ResourceGroup{}

	row := 0
	buffer := ""
	for _, v := range table {
		if !parseHead {
			if v == TBL_HEAD_BASE {
				headCount++
			} else if v == TBL_HEAD_ROOT {
				rootCount++
				headCount++
			} else if v == '\n' && headCount == 3 {
				parseHead = true
				name = buffer
				buffer = ""
			} else if v == ' ' {
				//
			} else {
				buffer += string(v)
			}
			continue
		}
		
		isArrow, isHead := p.isArrowComponent(v)
		if len(buffer) == 0 && isArrow && !passArrow {
			if arrowCount > 0 {
				passArrow = true
			}
			if isHead {
				arrowCount++
			}
		} else if v == '\n' {
			if row == 0 {
				heads = p.parseHeaders(buffer)
			} else {
				row := p.parseRow(buffer, heads)
				items = append(items, row)
			}
			buffer = ""
			arrowCount = 0
			passArrow = false
			row++
		} else {
			buffer += string(v)
		}
	}

	if len(buffer) > 0 {
		row := p.parseRow(buffer, heads)
		items = append(items, row)
	}

	return newNexus(name, rootCount == 2, items)
}

func (p *csvtParser) parseHeaders(row string) []string {
	heads := []string{}

	buffer := ""
	for _, v := range row {
		if v != HEA_SEPARATOR {	
			buffer += string(v)
			continue
		}
		heads = append(heads, buffer)
		buffer = ""
	}

	if len(buffer) > 0 {
		heads = append(heads, buffer)
	}

	return heads
}

func (p *csvtParser) parseRow(row string, header []string) ResourceGroup {
	instance := p.categoryOf(row, len(header) != 0)
	var group interface{}
	switch instance {
	case MAP:
		group = p.parseMap(row)
	case ARR:
		group = p.parseArray(row)
	case STR:
		group = p.parseStructure(row)
	case OBJ:
		group = p.parseObject(row)
	default:
		panic("//TODO: Not recognized.")
	}
	return newGroup(instance, header, group)
}

func (p *csvtParser) categoryOf(row string, header bool) GroupCategory {
	inString := false
	escape := false

	for _, v := range row {
		if !inString {
			if v == '"' {
				inString = true
			} else if v == MAP_LINKER {
				return MAP
			} else if v == ARR_SEPARATOR || v == ARR_CLOSING {
				return ARR
			} else if v == STR_SEPARATOR || v == STR_CLOSING {
				return STR
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
		return OBJ
	}

	return STR
}

func (p *csvtParser) parseMap(row string) map[string]ResourceNode {
	mapp := map[string]ResourceNode{}

	inString := false
	escape := false

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
				node := p.parseObject(buffer)
				key = node.key()
				buffer = ""
			} else if v == MAP_SEPARATOR {
				node := p.parseObject(buffer)
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

	node := p.parseObject(buffer)
	mapp[key] = node

	return mapp
}

func (p *csvtParser) parseArray(row string) []ResourceNode {
	return p.parseList(row, ARR_SEPARATOR, ARR_CLOSING)
}

func (p *csvtParser) parseStructure(row string) []ResourceNode {
	return p.parseList(row, STR_SEPARATOR, STR_CLOSING)
}

func (p *csvtParser) parseList(row string, separator, closing rune) []ResourceNode {
	lst := []ResourceNode{}

	inString := false
	escape := false

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
				node := p.parseObject(buffer)
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

	if len(buffer) > 0 {
		node := p.parseObject(buffer)
		lst = append(lst, node)
	}

	return lst
}

func (p *csvtParser) parseObject(obj string) ResourceNode {
	if len(obj) == 0 {
		return fromEmpty()
	}
	if v, i, ok := p.isPointer(obj); ok {
		return fromPointer(v, i)
	}
	if v, ok := p.isString(obj); ok {
		return fromNonPointer(v)
	}
	if v, err := strconv.ParseBool(obj); err == nil {
		return fromNonPointer(v)
	}
	if strings.Contains(obj, ".") {
		if v, err := strconv.ParseFloat(obj, 64); err == nil {
			return fromNonPointer(v)
		}
	}
	if ri, err := strconv.Atoi(obj); err == nil {
		return fromNonPointer(ri)
	}

	panic("//TODO: Does not recognized")
}

func (p *csvtParser) isPointer(obj string) (string, int, bool) {
	if obj[0] != byte(PTR_HEADER) {
		return obj, 0, false
	}

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

	return key, index, true
}

func (p *csvtParser) isString(obj string) (string, bool) {
	len := len(obj)
	if obj[0] == '"' && obj[len-1] == '"' {
		return obj[1 : len-1], true
	}
	return obj, false
}

func (p *csvtParser) isArrowComponent(v rune) (bool, bool) {
	return (v >= '0' && v <= '9') || v == TBL_INDEX_HEAD || v == '-' || v == '>' || v == ' ', v == '>'
}
