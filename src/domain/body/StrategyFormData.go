package body

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
)

func applyFormData(b *Body) *bytes.Buffer {
	var body *bytes.Buffer

	writer := multipart.NewWriter(body)
	for k, v := range b.Parameters {
		if k == DOCUMENT_PARAM {
			continue
		}

		if !v.IsFile {
			err := writer.WriteField(k, v.Value)
			if err != nil {
				//TODO: Log
			}
		} else {
			err := makeFormDataFile(&v, writer)
			if err != nil {
				//TODO: Log
			}
		}

	}

	writer.Close()
	return body
}

func makeFormDataFile(parameter *BodyParameter, writer *multipart.Writer) error {
	fileData, err := base64.StdEncoding.DecodeString(parameter.Value)
	if err != nil {
		return err
	}

	filePart, err := writer.CreateFormFile("file", parameter.FileName)
	if err != nil {
		return err
	}

	_, err = io.Copy(filePart, bytes.NewReader(fileData))
	if err != nil {
		return err
	}

	return nil
}
