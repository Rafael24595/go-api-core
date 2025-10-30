package test_utils

import (
	"reflect"
	"testing"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

func TestSplitByRune_Simple(t *testing.T) {
	input := "one.two.three"
	fragments := utils.SplitByRune(input, '.')

	expected := []string{"one", "two", "three"}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_NoSplitRune(t *testing.T) {
	input := "hello_golang"
	fragments := utils.SplitByRune(input, '.')
	
	expected := []string{"hello_golang"}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_LeadingSplitRune(t *testing.T) {
	input := ".start.middle.end"
	fragments := utils.SplitByRune(input, '.')

	expected := []string{"", "start", "middle", "end"}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_OnlyDots(t *testing.T) {
	input := "..."
	fragments := utils.SplitByRune(input, '.')

	expected := []string{"", "", "", ""}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_SingleEscaped(t *testing.T) {
	input := `one\.two.three`
	fragments := utils.SplitByRune(input, '.')

	expected := []string{`one.two`, "three"}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_EscapedAtEnd(t *testing.T) {
	input := `one.two\.`
	fragments := utils.SplitByRune(input, '.')

	expected := []string{"one", `two.`}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_DoubleBackslash(t *testing.T) {
	input := `one\\.two`
	fragments := utils.SplitByRune(input, '.')

	expected := []string{`one\`, "two"}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_MixedEscapes(t *testing.T) {
	input := `one.two\.three\\.four`
	fragments := utils.SplitByRune(input, '.')

	expected := []string{"one", `two.three\`, "four"}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}

func TestSplitByRune_MultipleBackslashes(t *testing.T) {
	input := `one\\\\\.two.three`
	fragments := utils.SplitByRune(input, '.')

	expected := []string{`one\\.two`, "three"}
	if !reflect.DeepEqual(fragments, expected) {
		t.Errorf("Found %#v, but %#v expected", fragments, expected)
	}
}
