package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
)

type CsvSerializer struct {
	Structures map[string][]string
}

func (s *CsvSerializer) Serialize(value any) {
	val := reflect.ValueOf(value)
	typ := val.Type().Name()

	s.serialize(value)

	t := s.key(val)

	keys := collection.FromMap(s.Structures).Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] == t
	})

	for _, k := range keys {
		v := s.Structures[k]
		pattern := "///"
		if k == t {
			pattern = "/**"
		}
		fmt.Printf("\n%s %s\n", pattern, k)
		for i, e := range v {
			index := strconv.FormatInt(int64(i - 1), 10)
			if i == 0 {
				index = "H"
			}
			fmt.Printf("%v-> %s\n", index, e)
		}
	}

	fmt.Printf("%v", typ)
}

func (s *CsvSerializer) serialize(value any) string {
	val := reflect.ValueOf(value)

	row := ""

	switch val.Kind() {
	case reflect.Struct:
		strRow := []string{}
		for i := 0; i < val.NumField(); i++ {
			value := val.Field(i).Interface()
			if !isCommonType(value) {
				value = s.serialize(value)
			} else {
				value = sprintf("%v", value)
			}

			strRow = append(strRow, fmt.Sprintf("%v", value))
		}
		row = fmt.Sprintf("%v%c", strings.Join(strRow, string(STR_SEPARATOR)), STR_CLOSING)
	case reflect.Map:
		mapRow := []string{}
		for _, k := range val.MapKeys() {
			key := k.Interface()
			if !isCommonType(key) {
				key = s.serialize(key)
			} else {
				key = sprintf("%v", key)
			}
				
			value := val.MapIndex(k).Interface()
			if !isCommonType(value) {
				value = s.serialize(value)
			} else {
				value = sprintf("%v", value)
			}
			mapRow = append(mapRow, fmt.Sprintf("%v%c%v", key, MAP_LINKER, value))
		}
		row = fmt.Sprintf("%v", strings.Join(mapRow, string(MAP_SEPARATOR)))
	case reflect.Slice, reflect.Array:
		arrayRow := []string{}
		for i := 0; i < val.Len(); i++ {
			value := val.Index(i).Interface()
			if !isCommonType(value) {
				value = s.serialize(value)
			}
			arrayRow = append(arrayRow, sprintf("%v", value))
		}
		row = fmt.Sprintf("%v%c", strings.Join(arrayRow, string(ARR_SEPARATOR)), ARR_CLOSING)
	case reflect.String:
		row = sprintf("\"%v\"", value)
	default:
		row = sprintf("%v", value)
	}

	t := s.key(val)

	_, exists := s.Structures[t]
	if !exists {
		headers, _ := s.headers(value)
		s.Structures[t] = append(s.Structures[t], headers)
	}
	s.Structures[t] = append(s.Structures[t], row)

	return fmt.Sprintf("$%s_%v", t, len(s.Structures[t])-2)
}

func (s *CsvSerializer) key(val reflect.Value) string {
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

func (s *CsvSerializer) headers(value any) (string, bool) {
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

	return strings.Join(headers, ";"), true
}

func (s CsvSerializer) sha1Identifier(input string) string {
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
