package openapi

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func MakeFromYaml(file []byte) (*OpenAPI, error) {
	raw, err := DeserializeFromYaml(file)
	if err != nil {
		return nil, err
	}

	return makeFromInterface(raw)
}

func DeserializeFromYaml(file []byte) (map[string]any, error) {
	var raw map[string]any
	err := yaml.Unmarshal(file, &raw)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func MakeFromJson(file []byte) (*OpenAPI, error) {
	var raw map[string]any
	err := json.Unmarshal(file, &raw)
	if err != nil {
		return nil, err
	}

	return makeFromInterface(raw)
}

func makeFromInterface(raw map[string]any) (*OpenAPI, error) {
	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var openAPI OpenAPI
	err = json.Unmarshal(jsonData, &openAPI)
	if err != nil {
		return nil, err
	}

	return &openAPI, nil
}