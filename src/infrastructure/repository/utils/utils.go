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

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	result, readErr  := io.ReadAll(file)
	err = file.Close()
	if err != nil {
		return make([]byte, 0), err
	}

	return result, readErr 
}

func WriteFile(filePath, content string) error {
	dir := filepath.Dir(filePath)
    err := os.MkdirAll(dir, os.ModePerm)
    if err != nil {
        return err
    }
	
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, errWrite := file.Write([]byte(content))
	err = file.Close()
	if err != nil {
		return err
	}

	if errWrite != nil {
		return errWrite
	}

	return nil
}
