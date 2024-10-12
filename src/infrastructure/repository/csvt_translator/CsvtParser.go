package csvt_translator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type csvtParser struct {
}

func newDeserializerParser() *csvtParser {
	return &csvtParser{
	}
}

func (p *csvtParser) parse(table string) (*ResourceNexus, TranslateError) {
	root := false
	
	items := []ResourceGroup{}

	fragments := strings.Split(table, "\n")

	if strings.Contains(fragments[0], string(TBL_HEAD_ROOT) + string(TBL_HEAD_ROOT)) {
		root = true
	}

	re := regexp.MustCompile(`/\*\*\s|///\s`)
	name := re.ReplaceAllString(fragments[0], "")

	heads := p.parseHeaders(fragments[1])

	for _, v := range fragments[1:] {
		row, err := p.parseRow(v, heads)
		if err != nil {
			return nil, err
		}
		items = append(items, *row)
	}
	
	result := newNexus(name, root, items)

	return &result, nil
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

func (p *csvtParser) parseRow(row string, header []string) (*ResourceGroup, TranslateError) {
	instance := p.categoryOf(row, len(header) != 0)
	var group interface{}
	var err TranslateError
	switch instance {
	case MAP:
		group, err = p.parseMap(row)
	case ARR:
		group, err = p.parseArray(row)
	case STR:
		group, err = p.parseStructure(row)
	case OBJ:
		group, err = p.parseObject(row)
	default:
		message := fmt.Sprintf("Row type not recognized: \n%s.", row)
		err = TranslateErrorFrom(message)
	}

	if err != nil {
		return nil, err
	}

	result := newGroup(instance, header, group)
	return &result, nil
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

func (p *csvtParser) parseMap(row string) (map[string]ResourceNode, TranslateError) {
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
				node, err := p.parseObject(buffer)
				if err != nil {
					return nil, err
				}
				key = node.key()
				buffer = ""
			} else if v == MAP_SEPARATOR {
				node, err := p.parseObject(buffer)
				if err != nil {
					return nil, err
				}
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

	node, err := p.parseObject(buffer)
	if err != nil {
		return nil, err
	}
	mapp[key] = node

	return mapp, nil
}

func (p *csvtParser) parseArray(row string) ([]ResourceNode, TranslateError) {
	return p.parseList(row, ARR_SEPARATOR, ARR_CLOSING)
}

func (p *csvtParser) parseStructure(row string) ([]ResourceNode, TranslateError) {
	return p.parseList(row, STR_SEPARATOR, STR_CLOSING)
}

func (p *csvtParser) parseList(row string, separator, closing rune) ([]ResourceNode, TranslateError) {
	lst := []ResourceNode{}
	
	for _, v := range strings.Split(row, string(separator)) {
		re := regexp.MustCompile(`\d+-> `)
		v := re.ReplaceAllString(v, "")

		v = strings.ReplaceAll(v, string(closing), "")

		node, err := p.parseObject(v)
		if err != nil {
			return nil, err
		}
		lst = append(lst, node)
	}

	return lst, nil
}

func (p *csvtParser) parseObject(obj string) (ResourceNode, TranslateError) {
	if len(obj) == 0 {
		return fromEmpty(), nil
	}
	if v, i, ok, err := p.isPointer(obj); ok {
		if err != nil {
			return ResourceNode{}, nil
		}
		return fromPointer(v, i), nil
	}
	if v, ok := p.isString(obj); ok {
		return fromNonPointer(v) , nil
	}
	if v, err := strconv.ParseBool(obj); err == nil { 
		return fromNonPointer(v) , nil
	}
	if strings.Contains(obj, ".") {
		if v, err := strconv.ParseFloat(obj, 64); err == nil {
			return fromNonPointer(v) , nil
		}
	}
	if v, err := strconv.Atoi(obj); err == nil {
		return fromNonPointer(v) , nil
	}

	err := fmt.Sprintf("Object type not recognized: \n%s.", obj)
	return ResourceNode{}, TranslateErrorFrom(err)
}

func (p *csvtParser) isPointer(obj string) (string, int, bool, TranslateError) {
	if obj[0] != byte(PTR_HEADER) {
		return obj, 0, false, nil
	}

	fragments := strings.Split(obj[1:], string(PTR_SEPARATOR))

	key := fragments[0]
	index := 0

	index, err := strconv.Atoi(fragments[1])
	if err != nil {
		message := fmt.Sprintf("Index \"%s\" type not recognized.", fragments[1])
		return "", 0, false, TranslateErrorFromCause(message, err)
	}

	return key, index, true, nil
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
