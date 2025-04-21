package utils

import (
	"strconv"
	"strings"
)

type Any struct {
	item string
}

func AnyFrom(item string) *Any {
	return &Any{
		item: item,
	}
}

func (a Any) Bool() (bool, bool) {
	val, err := strconv.ParseBool(strings.ToLower(a.item))
	if err != nil {
		return false, false
	}
	return val, true
}

func (a Any) String() (string, bool) {
	return a.item, true
}

func (a Any) Int() (int, bool) {
	val, err := strconv.Atoi(a.item)
	if err != nil {
		return 0, false
	}
	return val, true
}

func (a Any) Int64() (int64, bool) {
	val, err := strconv.ParseInt(a.item, 10, 64)
	if err != nil {
		return 0, false
	}
	return val, true
}

func (a Any) Float32() (float32, bool) {
	val, err := strconv.ParseFloat(a.item, 32)
	if err != nil {
		return 0, false
	}
	return float32(val), true
}

func (a Any) Float64() (float64, bool) {
	val, err := strconv.ParseFloat(a.item, 64)
	if err != nil {
		return 0, false
	}
	return val, true
}
