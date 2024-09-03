package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
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
		root := ""
		if k == t {
			root = " *ROOT"
		}
		fmt.Printf("\n///%s %s\n", root, k)
		for i, e := range v {
			fmt.Printf("%v-> %s\n", i, e)
		}
	}

	fmt.Printf("%v", typ)
}

func (s *CsvSerializer) serialize(value any) string {
	val := reflect.ValueOf(value)

	row := []string{}

	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			value := val.Field(i).Interface()
			if !isCommonType(value) {
				value = s.serialize(value)
			}

			row = append(row, fmt.Sprintf("%v", value))
		}
	case reflect.Map:
		mapRow := []string{}
		for _, key := range val.MapKeys() {
			value := val.MapIndex(key).Interface()
			if !isCommonType(value) {
				value = s.serialize(value)
			}
			mapRow = append(mapRow, fmt.Sprintf("%s=%v", key, value))
		}
		row = append(row, fmt.Sprintf("%v", strings.Join(mapRow, "#")))
	case reflect.Slice, reflect.Array:
		arrayRow := []string{}
		for i := 0; i < val.Len(); i++ {
			value := val.Index(i).Interface()
			if !isCommonType(value) {
				value = s.serialize(value)
			}
			arrayRow = append(arrayRow, fmt.Sprintf("%v", value))
		}
		row = append(row, fmt.Sprintf("%v", strings.Join(arrayRow, ",")))
	default:
		row = append(row, fmt.Sprintf("%v", value))
	}

	t := s.key(val)

	_, exists := s.Structures[t]
	if !exists {
		headers, isStruct := s.headers(value)
		if isStruct {
			s.Structures[t] = append(s.Structures[t], headers)
		}
	}
	s.Structures[t] = append(s.Structures[t], strings.Join(row, ";"))

	return fmt.Sprintf("$%s_%v", t, len(s.Structures[t])-1)
}

func (s *CsvSerializer) key(val reflect.Value) string {
	switch val.Kind() {
	case reflect.Map:
		return "common_map"
	case reflect.Slice, reflect.Array:
		return "common_array"
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

func isCommonType(value interface{}) bool {
	switch value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return true
	default:
		return false
	}
}
