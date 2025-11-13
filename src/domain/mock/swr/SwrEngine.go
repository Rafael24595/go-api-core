package swr

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-collections/collection"
	"golang.org/x/net/html/charset"
)

/*********************************************************************************************
*                     SWR – The Sequential Waggon Requirement language.                      *
*   Because writing requirements should feel like driving a train through data structures.   *
*********************************************************************************************/

type logger func(string)

func defaltLogger(string) {}

type swrEngine struct {
	logger    logger
	cache     map[string]any
	payload   string
	arguments map[string]string
}

func newSwrEngine(payload string, arguments map[string]string) *swrEngine {
	return &swrEngine{
		logger:    defaltLogger,
		cache:     make(map[string]any),
		payload:   payload,
		arguments: arguments,
	}
}

// NewEngine returns a new SWR engine initialized with no payload and empty arguments.
// This function is typically used to build an engine via method chaining.
func NewEngine() *swrEngine {
	return newSwrEngine("", make(map[string]string))
}

// MatchRequirement iterates through a list of requirement expressions (reqs)
// and evaluates them sequentially against the provided payload and arguments.
// It returns the first matching requirement and a boolean indicating success.
func MatchRequirement(reqs []string, payload string, arguments map[string]string) (string, bool) {
	return newSwrEngine(payload, arguments).Evalue(reqs)
}

// Payload sets the payload data (as a string) to be evaluated by the engine.
// It supports JSON, XML, or plain text content.
func (f *swrEngine) Payload(payload string) *swrEngine {
	f.payload = payload
	return f
}

// Arguments sets the argument map that can be referenced within SWR expressions.
func (f *swrEngine) Arguments(arguments map[string]string) *swrEngine {
	f.arguments = arguments
	return f
}

// Logger sets a custom logger function for internal debug or error reporting.
func (f *swrEngine) Logger(logger logger) *swrEngine {
	f.logger = logger
	return f
}

// Evalue sequentially evaluates a list of SWR requirement expressions
// against the current engine state (payload and arguments).
// It returns the first expression that successfully matches, along with a boolean
// indicating whether any requirement was satisfied.
func (f *swrEngine) Evalue(reqs []string) (string, bool) {
	for _, v := range reqs {
		if f.evalue(v) {
			return v, true
		}
	}
	return "", false
}

func (f *swrEngine) evalue(req string) bool {
	fragments := collection.VectorFromList(utils.SplitByRune(req, '.'))
	result, ok := f.match(fragments, true)
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

func (f *swrEngine) match(fragments *collection.Vector[string], isMain bool) (any, bool) {
	root, ok := f.findRoot(fragments)
	if !ok {
		return nil, false
	}

	target := root
	for fragments.Size() > 0 {
		isLogical := f.isNextLogical(fragments)
		if !isMain && isLogical {
			return target, true
		}

		cursor, ok := fragments.Shift()
		if !ok {
			return nil, false
		}

		if f.isLogical(*cursor) {
			target, ok = f.operate(*cursor, target, fragments)
			if !ok {
				return nil, false
			}
			continue
		}

		if raw, ok := findValue(*cursor); ok {
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

func (f *swrEngine) operate(operation string, target any, fragments *collection.Vector[string]) (bool, bool) {
	source, ok := f.match(fragments, false)
	if !ok {
		return false, false
	}

	switch operation {
	case string(mock.StepOperatorEq):
		t := f.tryToString(target)
		s := f.tryToString(source)
		return eq(t, s), true
	case string(mock.StepOperatorNe):
		t := f.tryToString(target)
		s := f.tryToString(source)
		return !eq(t, s), true
	case string(mock.StepOperatorGt):
		t := f.tryToNumeric(target)
		s := f.tryToNumeric(source)
		return gt(t, s, false), true
	case string(mock.StepOperatorGte):
		t := f.tryToNumeric(target)
		s := f.tryToNumeric(source)
		return gt(t, s, true), true
	case string(mock.StepOperatorLt):
		t := f.tryToNumeric(target)
		s := f.tryToNumeric(source)
		return lt(t, s, false), true
	case string(mock.StepOperatorLte):
		t := f.tryToNumeric(target)
		s := f.tryToNumeric(source)
		return lt(t, s, true), true
	case string(mock.StepOperatorAnd):
		return and(target, source), true
	case string(mock.StepOperatorOr):
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

func (f *swrEngine) moveCursor(cursor string, target any) (any, bool) {
	if value, ok := findPosition(cursor); ok {
		return f.moveCursorOnVector(value, target)
	}

	return f.moveCursorOnMap(cursor, target)
}

func (f *swrEngine) moveCursorOnVector(cursor string, target any) (any, bool) {
	index, err := strconv.Atoi(cursor)
	if err != nil {
		f.logger(err.Error())
		return nil, false
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		f.logger(fmt.Sprintf("the value '%v' is not a valid slice", target))
		return nil, false
	}

	if index < 0 || index >= v.Len() {
		f.logger(fmt.Sprintf("the index %d is outside of the slice range %d", index, v.Len()))
		return nil, false
	}

	return v.Index(index).Interface(), true
}

func (f *swrEngine) moveCursorOnMap(cursor string, target any) (any, bool) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Map {
		f.logger(fmt.Sprintf("the value '%v' is not a valid map", target))
		return nil, false
	}

	k := reflect.ValueOf(cursor)
	val := v.MapIndex(k)
	if !val.IsValid() {
		f.logger(fmt.Sprintf("the key '%s' is not valid for the map '%v'", cursor, target))
		return nil, false
	}

	return val.Interface(), true
}

func (f *swrEngine) findRoot(fragments *collection.Vector[string]) (any, bool) {
	cursor, ok := fragments.Shift()
	if !ok {
		return nil, false
	}

	switch *cursor {
	case string(mock.StepInputPayload):
		return f.findPayload(fragments)
	case string(mock.StepInputArguments):
		return f.arguments, true
	default:
		return findValue(*cursor)
	}
}

func (f *swrEngine) findPayload(fragments *collection.Vector[string]) (any, bool) {
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
	case string(mock.StepFormatText):
		return f.payload, true
	case string(mock.StepFormatJson):
		var payload map[string]any
		result, ok = f.getJsonPayload(payload)
	case string(mock.StepFormatVecJson):
		var payload []map[string]any
		result, ok = f.getJsonPayload(payload)
	case string(mock.StepFormatXml):
		var payload map[string]any
		result, ok = f.getXmlPayload(payload)
	case string(mock.StepFormatVecXml):
		var payload []map[string]any
		result, ok = f.getXmlPayload(payload)
	}

	if ok {
		f.cache[*content] = result
	}

	return result, ok
}

func (f *swrEngine) getJsonPayload(payload any) (any, bool) {
	err := json.Unmarshal([]byte(f.payload), &payload)
	if err != nil {
		f.logger(err.Error())
		return nil, false
	}
	return payload, true
}

func (f *swrEngine) getXmlPayload(payload any) (any, bool) {
	decoder := xml.NewDecoder(bytes.NewReader([]byte(f.payload)))
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&payload); err != nil {
		f.logger(err.Error())
		return nil, false
	}
	return payload, true
}

func (f *swrEngine) tryToNumeric(item any) any {
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.String {
		return item
	}

	r, err := strconv.ParseFloat(v.String(), 64)
	if err != nil {
		return item
	}

	return r
}

func (f *swrEngine) tryToString(item any) any {
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

func (f *swrEngine) isLogical(cursor string) bool {
	return strings.HasPrefix(cursor, "$")
}

func (f *swrEngine) isNextLogical(fragments *collection.Vector[string]) bool {
	next, ok := fragments.First()
	if !ok {
		return false
	}
	return *next == string(mock.StepOperatorAnd) ||
		*next == string(mock.StepOperatorOr)
}
