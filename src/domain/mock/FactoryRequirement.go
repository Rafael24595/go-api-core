package mock

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
	"golang.org/x/net/html/charset"
)

type factoryRequirement struct {
	cache   map[string]any
	payload string
	headers map[string]string
}

func newFactoryRequirement(payload string, headers map[string]string) factoryRequirement {
	return factoryRequirement{
		cache:   make(map[string]any),
		payload: payload,
		headers: headers,
	}
}

func FindRequirement(keys []string, payload string, headers map[string]string) (string, bool) {
	instance := newFactoryRequirement(payload, headers)
	for _, v := range keys {
		if instance.evalue(v) {
			return v, true
		}
	}
	return "", false
}

func (f *factoryRequirement) evalue(req string) bool {
	fragments := collection.VectorFromList(utils.SplitByRune(req, '.'))
	result, ok := f.match(fragments)
	if !ok {
		return ok
	}

	v := reflect.ValueOf(result)
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	default:
		return true
	}
}

func (f *factoryRequirement) match(fragments *collection.Vector[string]) (any, bool) {
	root, ok := f.findRoot(fragments)
	if !ok {
		return nil, false
	}

	target := root
	for fragments.Size() > 0 {
		cursor, ok := fragments.Shift()
		if !ok {
			return nil, false
		}

		if strings.HasPrefix(*cursor, "$") {
			target, ok = f.operate(*cursor, target, fragments)
			if !ok {
				return nil, false
			}
			continue
		}

		if raw, ok := f.findRaw(*cursor); ok {
			target = raw
			continue
		}

		target, ok = f.moveCursor(*cursor, target)
		if !ok {
			return nil, false
		}
	}

	return target, true
}

func (f *factoryRequirement) operate(operation string, target any, fragments *collection.Vector[string]) (bool, bool) {
	source, ok := f.match(fragments)
	if !ok {
		return false, false
	}

	switch operation {
	case "$eq":
		t := tryToString(target)
		s := tryToString(source)
		return eq(t, s), true
	case "$ne":
		t := tryToString(target)
		s := tryToString(source)
		return !eq(t, s), true
	case "$gt":
		t := tryToNumeric(target)
		s := tryToNumeric(source)
		return gt(t, s, false), true
	case "$gte":
		t := tryToNumeric(target)
		s := tryToNumeric(source)
		return gt(t, s, true), true
	case "$lt":
		t := tryToNumeric(target)
		s := tryToNumeric(source)
		return lt(t, s, false), true
	case "$lte":
		t := tryToNumeric(target)
		s := tryToNumeric(source)
		return lt(t, s, true), true
	case "$and":
		return and(target, source), true
	case "$or":
		return or(target, source), true
	}

	return false, false
}

func eq(a, b any) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if va.Type() != vb.Type() {
		return false
	}

	if !va.Type().Comparable() {
		return false
	}

	return va.Interface() == vb.Interface()
}

func gt(a, b any, e bool) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if !va.IsValid() || !vb.IsValid() {
		return false
	}

	if va.Kind() != vb.Kind() {
		return false
	}

	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return va.Int() > vb.Int() || (e && va.Int() == vb.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return va.Uint() > vb.Uint() || (e && va.Uint() == vb.Uint())
	case reflect.Float32, reflect.Float64:
		return va.Float() > vb.Float() || (e && va.Float() == vb.Float())
	case reflect.String:
		return va.String() > vb.String() || (e && va.String() == vb.String())
	default:
		return false
	}
}

func lt(a, b any, e bool) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if !va.IsValid() || !vb.IsValid() {
		return false
	}

	if va.Kind() != vb.Kind() {
		return false
	}

	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return va.Int() < vb.Int() || (e && va.Int() == vb.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return va.Uint() < vb.Uint() || (e && va.Uint() == vb.Uint())
	case reflect.Float32, reflect.Float64:
		return va.Float() < vb.Float() || (e && va.Float() == vb.Float())
	case reflect.String:
		return va.String() < vb.String() || (e && va.String() == vb.String())
	default:
		return false
	}
}

func and(a, b any) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if !va.IsValid() || !vb.IsValid() {
		return false
	}

	if va.Kind() != vb.Kind() {
		return false
	}

	if va.Kind() != reflect.Bool && vb.Kind() != reflect.Bool {
		return false
	}

	return va.Bool() && vb.Bool()
}

func or(a, b any) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if !va.IsValid() || !vb.IsValid() {
		return false
	}

	if va.Kind() != vb.Kind() {
		return false
	}

	if va.Kind() != reflect.Bool && vb.Kind() != reflect.Bool {
		return false
	}

	return va.Bool() || vb.Bool()
}

func (f *factoryRequirement) moveCursor(cursor string, target any) (any, bool) {
	re := regexp.MustCompile(`^\[([0-9]+)\]$`)
	matches := re.FindStringSubmatch(cursor)
	if len(matches) == 2 {
		return f.moveCursorOnVector(matches[1], target)
	}

	return f.moveCursorOnMap(cursor, target)
}

func (f *factoryRequirement) moveCursorOnVector(cursor string, target any) (any, bool) {
	index, err := strconv.Atoi(cursor)
	if err != nil {
		//TODO: Add flag to log the error.
		return nil, false
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		//TODO: Add flag to log the error.
		return nil, false
	}

	if index < 0 || index >= v.Len() {
		//TODO: Add flag to log the error.
		return nil, false
	}

	return v.Index(index).Interface(), true
}

func (f *factoryRequirement) moveCursorOnMap(cursor string, target any) (any, bool) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Map {
		//TODO: Add flag to log the error.
		return nil, false
	}

	k := reflect.ValueOf(cursor)
	val := v.MapIndex(k)
	if !val.IsValid() {
		//TODO: Add flag to log the error.
		return nil, false
	}

	return val.Interface(), true
}

func (f *factoryRequirement) findRaw(cursor string) (any, bool) {
	re := regexp.MustCompile(`^<(.+)>$`)
	matches := re.FindStringSubmatch(cursor)
	if len(matches) < 2 {
		return nil, false
	}
	return matches[1], true
}

func (f *factoryRequirement) findRoot(fragments *collection.Vector[string]) (any, bool) {
	root, ok := fragments.Shift()
	if !ok {
		return nil, false
	}

	switch *root {
	case "header":
		return f.headers, true
	case "payload":
		return f.findPayload(fragments)
	}

	return f.findRaw(*root)
}

func (f *factoryRequirement) findPayload(fragments *collection.Vector[string]) (any, bool) {
	content, ok := fragments.Shift()
	if !ok {
		return nil, false
	}

	if cached, ok := f.cache[*content]; ok {
		return cached, ok
	}

	var result any
	ok = false
	switch *content {
	case "text":
		return f.payload, true
	case "json":
		var payload map[string]any
		result, ok = f.getJsonPayload(payload)
	case "vec_json":
		var payload []map[string]any
		result, ok = f.getJsonPayload(payload)
	case "xml":
		var payload map[string]any
		result, ok = f.getXmlPayload(payload)
	case "vec_xml":
		var payload []map[string]any
		result, ok = f.getXmlPayload(payload)
	}

	if ok {
		f.cache[*content] = result
	}

	return result, ok
}

func (f *factoryRequirement) getJsonPayload(payload any) (any, bool) {
	err := json.Unmarshal([]byte(f.payload), &payload)
	if err != nil {
		//TODO: Add flag to log the error.
		return nil, false
	}
	return payload, true
}

func (f *factoryRequirement) getXmlPayload(payload any) (any, bool) {
	decoder := xml.NewDecoder(bytes.NewReader([]byte(f.payload)))
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&payload); err != nil {
		//TODO: Add flag to log the error.
		return nil, false
	}
	return payload, true
}

func tryToNumeric(item any) any {
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.String {
		return item
	}

	f, err := strconv.ParseFloat(v.String(), 64)
	if err != nil {
		return item
	}

	return f
}

func tryToString(item any) any {
	v := reflect.ValueOf(item)

	switch v.Kind() {
	case reflect.String:
		return item
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		return item
	}
}
