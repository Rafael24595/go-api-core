package csvt_translator

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
)

const (
	HEADER_ROOT = string(TBL_HEAD_BASE) + string(TBL_HEAD_ROOT) + string(TBL_HEAD_ROOT)
	HEADER_REGULAR = string(TBL_HEAD_BASE) + string(TBL_HEAD_BASE) + string(TBL_HEAD_BASE)
	POINTER_INDEX_FIX = 2
)

type CsvtSerializer struct {
	tables map[string][]string
}

func NewSerializer() *CsvtSerializer {
	return &CsvtSerializer{
		tables: make(map[string][]string),
	}
}

func (s *CsvtSerializer) Serialize(entities ...any) string {
	if len(entities) == 0 {
		return ""
	}

	rootKey := s.key(reflect.ValueOf(entities[0]))
	for _, e := range entities {
		s.serialize(e)
	}

	return s.makeTables(rootKey)
}

func (s *CsvtSerializer) makeTables(rootKey string) string {
	keys := collection.FromMap(s.tables).
		KeysCollection().
		Sort(func(a, b string) bool {
			return a == rootKey
		})

	buffer := ""
	for _, k := range keys.Collect() {
		rows := s.tables[k]

		pattern := HEADER_REGULAR
		if k == rootKey {
			pattern = HEADER_ROOT
		}

		buffer += fmt.Sprintf("\n%s %s\n", pattern, k)
		buffer += s.makeTableRows(rows)
	}

	return buffer
}

func (s *CsvtSerializer) makeTableRows(rows []string) string {
	buffer := ""
	for i, r := range rows {
		index := strconv.FormatInt(int64(i - 1), 10)
		if i == 0 {
			index = string(TBL_INDEX_HEAD)
		}
		buffer += fmt.Sprintf("%s%s\n", s.formatIndexArrow(index), r)
	}

	return buffer
}

func (s *CsvtSerializer) serialize(entity any) string {
	rEntity := reflect.ValueOf(entity)

	key := s.key(rEntity)
	row := s.serializeEntity(entity, rEntity)

	if _, exists := s.tables[key]; !exists {
		headers, _ := s.headers(entity)
		s.tables[key] = append(s.tables[key], headers)
	}

	s.tables[key] = append(s.tables[key], row)

	return s.formatPointerReference(key, len(s.tables[key]))
}

func (s *CsvtSerializer) serializeEntity(entity any, rEntity reflect.Value) string {
	switch rEntity.Kind() {
	case reflect.Struct:
		return s.serializeStruct(rEntity)
	case reflect.Map:
		return s.serializeMap(rEntity)
	case reflect.Slice, reflect.Array:
		return s.serializeArray(rEntity)
	default:
		return s.serializeObject(entity, rEntity)
	}
}

func (s *CsvtSerializer) serializeStruct(entity reflect.Value) string {
	strRow := []string{}
	for i := 0; i < entity.NumField(); i++ {
		value := entity.Field(i).Interface()
		if !isCommonType(value) {
			value = s.serialize(value)
		} else {
			value = sprintf("%v", value)
		}

		strRow = append(strRow, fmt.Sprintf("%v", value))
	}
	return fmt.Sprintf("%v%c", strings.Join(strRow, string(STR_SEPARATOR)), STR_CLOSING)
}

func (s *CsvtSerializer) serializeMap(entity reflect.Value) string {
	mapRow := []string{}

	for _, k := range entity.MapKeys() {
		key := k.Interface()
		if !isCommonType(key) {
			key = s.serialize(key)
		} else {
			key = sprintf("%v", key)
		}
			
		value := entity.MapIndex(k).Interface()
		if !isCommonType(value) {
			value = s.serialize(value)
		} else {
			value = sprintf("%v", value)
		}
		mapRow = append(mapRow, fmt.Sprintf("%v%c%v", key, MAP_LINKER, value))
	}

	return fmt.Sprintf("%v", strings.Join(mapRow, string(MAP_SEPARATOR)))
}

func (s *CsvtSerializer) serializeArray(entity reflect.Value) string {
	arrayRow := []string{}
	for i := 0; i < entity.Len(); i++ {
		value := entity.Index(i).Interface()
		if !isCommonType(value) {
			value = s.serialize(value)
		}
		arrayRow = append(arrayRow, sprintf("%v", value))
	}
	return fmt.Sprintf("%v%c", strings.Join(arrayRow, string(ARR_SEPARATOR)), ARR_CLOSING)
}

func (s *CsvtSerializer) serializeObject(entity any, rEntity reflect.Value) string {
	if rEntity.Kind() == reflect.String {
		return sprintf("\"%v\"", entity)
	}
	return sprintf("%v", entity)
}

func (s *CsvtSerializer) key(val reflect.Value) string {
	switch val.Kind() {
	case reflect.Map:
		return "common-map"
	case reflect.Slice, reflect.Array:
		return "common-array"
	default:
		typ := val.Type()
		return fmt.Sprintf("%s&%s", typ.Name(), s.sha1Identifier(typ.PkgPath()))
	}
}

func (s *CsvtSerializer) headers(value any) (string, bool) {
	val := reflect.ValueOf(value)
	typ := val.Type()

	headers := []string{}

	if val.Kind() != reflect.Struct {
		return "", false
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i).Name
		csvTag := typ.Field(i).Tag.Get("csv")
		if csvTag != "" {
			field = csvTag
		}

		headers = append(headers, field)
	}

	return strings.Join(headers, string(HEA_SEPARATOR)), true
}

func (s *CsvtSerializer) formatIndexArrow(index string) string {
	return fmt.Sprintf("%v-> ", index)
}

func (s *CsvtSerializer) formatPointerReference(key string, position int) string {
	return fmt.Sprintf("$%s_%v", key, position - POINTER_INDEX_FIX)
}

func (s CsvtSerializer) sha1Identifier(input string) string {
	hash := sha1.New()
	hash.Write([]byte(input))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

func sprintf(pattern string, values ...any) string {
	for i, v := range values {
		switch v := v.(type) {
		case string:
			values[i] = fmt.Sprintf("\"%v\"", v)
		}
	}
	return fmt.Sprintf(pattern, values...)
}

func isCommonType(value interface{}) bool {
	switch value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return true
	default:
		return false
	}
}
