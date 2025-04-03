package utils

import (
	"encoding/json"
	"strings"
)

func FindJson[T any](path string, raw map[string]any, item *T) error {
	fragments := strings.Split(path, ".")

	if len(fragments) > 0 && len(raw) == 0 {
		return nil
	}

	var level any
	level = raw
	for _, v := range fragments {
		valide, ok := level.(map[string]any)
		if !ok {
			return nil
		}

		level, ok = valide[v]
		if !ok {
			return nil
		}
	}

	jsonData, err := json.Marshal(level)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, item)
	if err != nil {
		return err
	}

	return nil
}
