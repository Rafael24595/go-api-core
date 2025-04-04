package test_openapi

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/openapi"
)

func TestMakeFromRequestBody(t *testing.T) {
	file, err := os.Open("test_openaoi.yaml")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	yaml, err := io.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	oapi, err := openapi.MakeFromYaml(yaml)
	if err != nil {
		t.Error(err)
	}

	raw, err := openapi.DeserializeFromYaml(yaml)
	if err != nil {
		t.Error(err)
	}

	builder := collection.NewBuilderCollection("anonymoys", oapi).SetRaw(raw)

	payload := oapi.Paths["/request"].Post.RequestBody
	result := builder.MakeFromRequestBody(payload)

	var request map[string]interface{}
	if err := json.Unmarshal(result.Payload, &request); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	value := request["method"]
	expected := "Request method"
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}

	value = request["uri"]
	expected = "Request Uri"
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}
}

func TestMakeFromSchema(t *testing.T) {
	file, err := os.Open("test_openaoi.yaml")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	yaml, err := io.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	oapi, err := openapi.MakeFromYaml(yaml)
	if err != nil {
		t.Error(err)
	}

	raw, err := openapi.DeserializeFromYaml(yaml)
	if err != nil {
		t.Error(err)
	}

	builder := collection.NewBuilderCollection("anonymoys", oapi).SetRaw(raw)

	schema := oapi.Paths["/collection/{userId}"].Get.Responses["200"].Content["application/json"].Schema
	example := builder.MakeFromSchema(&schema)

	var requests []map[string]interface{}
	if err := json.Unmarshal([]byte(example), &requests); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request object, got %v", requests)
	}

	request := requests[0]

	value := request["id"]
	expected := "000A1"
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}

	context, ok := request["context"].(map[string]interface{})
	if !ok {
		t.Errorf("Context not found or incorrect format.")
	}

	value = context["status"]
	expected = "pending"
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}

	reqs, ok := request["requests"].([]interface{})
	if !ok || len(reqs) != 1 {
		t.Fatalf("Expected 1 request object, got %v", reqs)
	}

	req := reqs[0].(map[string]interface{})

	value = req["method"]
	expected = "Request method"
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}

	fValue := req["timestamp"].(float64)
	fExpected := float64(1743433941068)
	if fValue != fExpected {
		t.Errorf("Found variable %v but %v expected", fValue, fExpected)
	}
}
