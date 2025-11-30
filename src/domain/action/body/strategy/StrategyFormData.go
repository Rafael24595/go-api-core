package body_strategy

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
)

const (
	FORM_DATA_PARAM = "form-data"
)

func FormDataBody(status bool, contentType domain.ContentType, builder *BuilderFormDataBody) *body.BodyRequest {
	parameters := make(map[string]map[string][]body.BodyParameter)
	parameters[FORM_DATA_PARAM] = builder.formData

	return body.NewBody(status, contentType, parameters)
}

type BuilderFormDataBody struct {
	formData map[string][]body.BodyParameter
}

func NewBuilderFromDataBody() *BuilderFormDataBody {
	return &BuilderFormDataBody{
		formData: make(map[string][]body.BodyParameter),
	}
}

func (b *BuilderFormDataBody) Add(key string, parameter *body.BodyParameter) *BuilderFormDataBody {
	var parameters []body.BodyParameter
	if exists, ok := b.formData[key]; ok {
		parameters = exists
	} else {
		parameters = make([]body.BodyParameter, 0)
	}

	b.formData[key] = append(parameters, *parameter)

	return b
}

func FindFormDataParameterIndex(b *body.BodyRequest, key string, index int) (*body.BodyParameter, bool) {
	if b.Parameters == nil {
		return nil, false
	}

	form, ok := b.Parameters[FORM_DATA_PARAM]
	if !ok {
		return nil, false
	}

	field, ok := form[key]
	if !ok {
		return nil, false
	}

	if index < 0 || index > len(field) {
		return nil, false
	}

	return &field[index], true
}

func AddFormData(b *body.BodyRequest, key string, parameter *body.BodyParameter) *body.BodyRequest {
	if b.Parameters == nil {
		b.Parameters = make(map[string]map[string][]body.BodyParameter)
	}

	form, ok := b.Parameters[FORM_DATA_PARAM]
	if !ok {
		form = make(map[string][]body.BodyParameter)
	}

	field, ok := form[key]
	if !ok {
		field = make([]body.BodyParameter, 0)
	}

	parameter.Order = int64(len(field) + 1)

	form[key] = append(field, *parameter)
	b.Parameters[FORM_DATA_PARAM] = form
	b.ContentType = domain.Form

	return b
}

func applyFormData(b *body.BodyRequest, q *query.Queries) (*bytes.Buffer, *query.Queries) {
	if !hasFiles(b) {
		return applyFormEncode(b, q)
	}

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	for k, p := range b.Parameters[FORM_DATA_PARAM] {
		for _, v := range p {
			if !v.Status {
				continue
			}

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

	err := writer.Close()
	if err != nil {
		log.Error(err)
	}

	return body, q
}

func makeFormDataFile(parameter *body.BodyParameter, writer *multipart.Writer) error {
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
