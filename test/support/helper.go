package support_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func ReadText(t *testing.T, filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	if err := file.Close(); err != nil {
		t.Error(err)
	}

	return bytes
}

func ReadJSON[T any](t *testing.T, filename string) T {
	file, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}

	if err := file.Close(); err != nil {
		t.Error(err)
	}

	bytes := ReadText(t, filename)

	var payload T
	err = json.Unmarshal(bytes, &payload)
	if err != nil {
		t.Error(err)
	}

	return payload
}
