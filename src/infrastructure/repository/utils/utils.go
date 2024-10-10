package utils

import (
	"io"
	"os"
	"path/filepath"
)

func ReadFile(filePath string) ([]byte, error) {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return make([]byte, 0), err
	}

	file, err := os.Open(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return make([]byte, 0), err
		}
		file, err = os.Create(filePath)
		if err != nil {
			return make([]byte, 0), err
		}
		defer file.Close()
		return make([]byte, 0), err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func WriteFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}