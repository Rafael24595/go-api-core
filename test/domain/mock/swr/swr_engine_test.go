package swr_test

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	support_test "github.com/Rafael24595/go-api-core/test/support"
)

func TestMatchRequirement_EQ_String(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$eq.<zig>"
	key2 := "payload.json.lang.[0].code.$eq.<golang>"

	keys := []string{key1, key2}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key2
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_EQ_Integer(t *testing.T) {
	key1 := "payload.json.lang.[0].order.$eq.<3>"
	key2 := "payload.json.lang.[2].order.$eq.<3>"

	keys := []string{key1, key2}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key2
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_EQ_Float(t *testing.T) {
	key1 := "payload.json.lang.[0].rate.$eq.<1.0>"
	key2 := `payload.json.lang.[2].rate.$eq.<0\.175>`

	keys := []string{key1, key2}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key2
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_EQ_Boolean(t *testing.T) {
	key1 := "payload.json.lang.[0].active.$eq.<true>"
	key2 := "payload.json.lang.[1].order.$eq.<true>"

	keys := []string{key1, key2}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_EQ_WithCorrectSourceReference(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$eq.payload.json.lang.[0].code"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_EQ_WithIncorrectSourceReference(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$eq.payload.json.lang.[1].code"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_EQ_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$eq.<zig>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_NE(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$ne.<zig>"
	key2 := "payload.json.lang.[0].code.$eq.<rust>"
	key3 := "payload.json.lang.[0].code.$ne.<rust>"

	keys := []string{key1, key2, key3}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_NE_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$eq.<zig>"
	key2 := "payload.json.lang.[0].code.$eq.<rust>"

	keys := []string{key1, key2}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_GT(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$gt.<1>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_GT_Negative(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$gt.<-11>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_GT_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$gt.<3>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_GTE(t *testing.T) {
	key1 := `payload.json.lang.[2].rate.$gte.<0\.10>`

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_GTE_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$gt.<3>"
	key2 := "payload.json.lang.[0].order.$gte.<11>"

	keys := []string{key1, key2}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_LT(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$lt.<11>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_LT_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$lt.<0>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_LT_Negative(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$lt.<-11>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_LTE(t *testing.T) {
	key1 := `payload.json.lang.[2].order.$lte.<3\.11>`

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_LTE_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[2].order.$lte.<1>"
	key2 := "payload.json.lang.[0].order.$lte.<0>"

	keys := []string{key1, key2}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_AND(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$eq.<golang>.$and.payload.json.lang.[0].order.$eq.<1>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_AND_Long(t *testing.T) {
	key1 := "payload.json.lang.[0].code.$eq.<golang>.$and.payload.json.lang.[0].order.$eq.<1>"
	key1 += ".$and.payload.json.lang.[2].code.$eq.<rust-lang>.$and.payload.json.lang.[2].order.$eq.<3>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_AND_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[1].code.$eq.<golang>.$and.payload.json.lang.[0].order.$eq.<1>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_OR(t *testing.T) {
	key1 := "payload.json.lang.[2].code.$eq.<rust-lang>.$or.payload.json.lang.[0].order.$eq.<1>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_OR_NotFound(t *testing.T) {
	key1 := "payload.json.lang.[1].code.$eq.<rust-lang>.$or.payload.json.lang.[2].order.$eq.<1>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}

func TestMatchRequirement_OR_Long(t *testing.T) {
	key1 := "payload.json.lang.[2].code.$eq.<zig-lang>.$or.payload.json.lang.[0].order.$ne.<1>"
	key1 += ".$or.payload.json.lang.[1].code.$eq.<rust-lang>.$or.payload.json.lang.[2].order.$eq.<1>"
	key1 += `.$or.payload.json.lang.[0].rate.$eq.<1\.0>.$or.payload.json.lang.[2].active.$eq.<false>`

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_OR_WithCorrectReferenceSource(t *testing.T) {
	key1 := "payload.json.lang.[2].code.$eq.<zig-lang>.$or.payload.json.lang.[0].order.$ne.<1>"
	key1 += `.$or.payload.json.lang.[0].rate.$eq.<1\.0>.$or.payload.json.lang.[2].active.$eq.payload.json.lang.[2].active`
	key1 += ".$or.payload.json.lang.[1].code.$eq.<rust-lang>.$or.payload.json.lang.[2].order.$eq.<1>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, _ := swr.MatchRequirement(keys, string(payload), headers)

	expected := key1
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestMatchRequirement_OR_WithIncorrectReferenceSource(t *testing.T) {
	key1 := "payload.json.lang.[2].code.$eq.<zig-lang>.$or.payload.json.lang.[0].order.$ne.<1>"
	key1 += `.$or.payload.json.lang.[0].rate.$eq.<1\.0>.$or.payload.json.lang.[2].active.$eq.payload.json.lang.[0].active`
	key1 += ".$or.payload.json.lang.[1].code.$eq.<rust-lang>.$or.payload.json.lang.[2].order.$eq.<1>"

	keys := []string{key1}
	payload := support_test.ReadText(t, "../../../support/test_source_langs.json")
	headers := make(map[string]string)

	result, ok := swr.MatchRequirement(keys, string(payload), headers)
	if ok {
		t.Errorf("Result found %#v, but nothing expected", result)
	}
}
