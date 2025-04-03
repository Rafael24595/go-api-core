package test_utils

import (
	"io"
	"os"
	"testing"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain/openapi"
)

func readRaw(t *testing.T) map[string]any {
	file, err := os.Open("test_openaoi.yaml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	yamlFile, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	raw, err := openapi.DeserializeFromYaml(yamlFile)
	if err != nil {
		t.Error(err)
	}

	return raw
}

func TestFindJson(t *testing.T) {
	raw := readRaw(t)

	path := "components.schemas.Task"
	var schema *openapi.Schema

	err := utils.FindJson(path, raw, &schema)
	if err != nil {
		t.Error(err)
	}

	if schema == nil {
		t.Errorf("Json path '%s' not found.", path)
	}

	expected := "object"
	if schema.Type != expected {
		t.Errorf("Found variable %s but %s expected", schema.Type, expected)
	}
}

func TestFindJsonFail(t *testing.T) {
	raw := readRaw(t)

	path := "components.schemas.Undefined"
	var schema *openapi.Schema

	err := utils.FindJson(path, raw, &schema)
	if err != nil {
		t.Error(err)
	}
	
	if schema != nil {
		t.Errorf("Undefined Json path '%s' found.", path)
	}
}
