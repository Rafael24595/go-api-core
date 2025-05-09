package body

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"

	"github.com/Rafael24595/go-api-core/src/commons/log"
)

const (
	FORM_DATA_PARAM = "form-data"
)

func applyFormData(b *BodyRequest) *bytes.Buffer {
	var body *bytes.Buffer

	writer := multipart.NewWriter(body)
	for k, v := range b.Parameters[FORM_DATA_PARAM] {
		for _, v := range v {
			if !v.IsFile {
				err := writer.WriteField(k, v.Value)
				if err != nil {
					log.Error(err)
				}
			} else {
				err := makeFormDataFile(&v, writer)
				if err != nil {
					log.Error(err)
				}
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
