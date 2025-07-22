package test_utils

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

type testCaseVersion struct {
	Input       string         `json:"input"`
	Expected    *utils.Version `json:"expected"`
	ExpectError bool           `json:"expectError"`
}

func TestParseVersion(t *testing.T) {
	file, err := os.Open("version_test_cases.json")
	if err != nil {
		t.Fatalf("Failed to open test case file: %v", err)
	}

	var testCases []testCaseVersion
	if err := json.NewDecoder(file).Decode(&testCases); err != nil {
		t.Fatalf("Failed to decode test case JSON: %v", err)
	}

	if err = file.Close(); err != nil {
		t.Error(err)
	}

	for _, test := range testCases {
		result, err := utils.ParseVersion(test.Input)
		if test.ExpectError {
			if err == nil {
				t.Errorf("Expected error for input %q, got nil", test.Input)
			}
			continue
		}

		if err != nil {
			t.Errorf("Did not expect error for input %q, got: %v", test.Input, err)
		}

		if !reflect.DeepEqual(result, test.Expected) {
			t.Errorf("For input %q, expected %+v, got %+v", test.Input, test.Expected, result)
		}
	}
}
