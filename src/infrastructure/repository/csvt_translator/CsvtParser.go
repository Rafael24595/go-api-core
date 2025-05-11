package csvt_translator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/log"
)

type csvtParser struct {
}

func newDeserializerParser() *csvtParser {
	return &csvtParser{}
}

func (p *csvtParser) parse(table string) (*ResourceNexus, TranslateError) {
	root := false

	items := []ResourceGroup{}

	fragments := strings.Split(table, "\n")

	if strings.Contains(fragments[0], string(TBL_HEAD_ROOT)+string(TBL_HEAD_ROOT)) {
		root = true
	}

	re := regexp.MustCompile(`/\*\*\s|///\s`)
	name := re.ReplaceAllString(fragments[0], "")

	heads := p.parseHeaders(fragments[1])

	for _, v := range fragments[2:] {
		if len(v) == 0 {
			continue
		}
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
	re := regexp.MustCompile(`[A-Za-z0-9]+->\s`)
	row = re.ReplaceAllString(row, "")
	if row == "" {
		return []string{}
	}
	return strings.Split(row, string(HEA_SEPARATOR))
}

func (p *csvtParser) parseRow(row string, header []string) (*ResourceGroup, TranslateError) {
	re := regexp.MustCompile(`\d+-> `)
	row = re.ReplaceAllString(row, "")

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
	switch rune(row[len(row)-1]) {
	case ARR_CLOSING:
		return ARR
	case MAP_CLOSING:
		return MAP
	case STR_CLOSING:
		return STR
	}

	if !header {
		return OBJ
	}

	return STR
}

func (p *csvtParser) parseMap(row string) (map[string]ResourceNode, TranslateError) {
	mapp := map[string]ResourceNode{}

	if rune(row[len(row)-1]) != MAP_CLOSING {
		//TODO: Expand description
		log.Panics("Unsupported")
	}

	row = row[:len(row)-1]

	buffer := row
	for len(buffer) > 0 {
		var index int
		if buffer[0] == '"' {
			index = strings.Index(buffer[1:], "\"") + 2
		} else {
			index = strings.Index(buffer, string(MAP_LINKER))
		}

		if index == -1 {
			//TODO: Expand description
			log.Panics("Value undefined")
		}

		key := buffer[:index]
		buffer = buffer[index+1:]

		node, err := p.parseObject(key)
		if err != nil {
			return nil, err
		}
		key = node.key()

		if buffer[0] == '"' {
			index = strings.Index(buffer[1:], "\"") + 1
			if (index < len(buffer) - 1 && buffer[index+1] == byte(MAP_SEPARATOR)) {
				index = index + 1
			} else {
				index = -1
			}
		} else {
			index = strings.Index(buffer, string(MAP_SEPARATOR))
		}

		var content string
		if index != -1 {
			if len(buffer) >= index && rune(buffer[index]) != MAP_SEPARATOR {
				//TODO: Expand description
				log.Panics("Unsupported")
			}
			content = buffer[:index]
			buffer = buffer[index+1:]
		} else {
			content = buffer
			buffer = ""
		}

		node, err = p.parseObject(content)
		if err != nil {
			return nil, err
		}
		mapp[key] = node
	}

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

	if rune(row[len(row)-1]) != closing {
		//TODO: Expand description
		log.Panics("Unsupported")
	}

	row = row[:len(row)-1]

	buffer := row
	for len(buffer) > 0 {
		var index int
		if buffer[0] == '"' {
			index = strings.Index(buffer[1:], "\"") + 2
			if len(buffer) == index {
				index = -1
			}
		} else {
			index = strings.Index(buffer, string(separator))
		}

		var content string
		if index != -1 {
			if len(buffer) >= index && rune(buffer[index]) != separator {
				//TODO: Expand description
				log.Panics("Unsupported")
			}
			content = buffer[:index]
			buffer = buffer[index+1:]
		} else {
			content = buffer
			buffer = ""
		}

		node, err := p.parseObject(content)
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
		return fromNonPointer(v), nil
	}
	lower := strings.ToLower(obj) 
	if lower == "false" {
		return fromNonPointer(false), nil
	}
	if lower == "true" {
		return fromNonPointer(true), nil
	}
	if strings.Contains(obj, ".") {
		if v, err := strconv.ParseFloat(obj, 64); err == nil {
			return fromNonPointer(v), nil
		}
	}
	if v, err := strconv.Atoi(obj); err == nil {
		return fromNonPointer(v), nil
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
		fixed := strings.ReplaceAll(obj[1 : len-1], "\\'", "\"")
		replacer := strings.NewReplacer(
			"\\\\", "\\",
			"\\\"", "\"",
			"\\n", "\n",
			"\\r", "\r",
			"\\t", "\t",
			"\\b", "\b",
			"\\f", "\f")
		fixed = replacer.Replace(fixed)
		return fixed, true
	}
	return obj, false
}
