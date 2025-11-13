package mock

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
)

type Response struct {
	Status    int      `json:"status"`
	Condition string   `json:"condition"`
	Name      string   `json:"name"`
	Headers   []Header `json:"headers"`
	Body      Body     `json:"body"`
}

type Body struct {
	ContentType domain.ContentType `json:"content_type"`
	Payload     string             `json:"string"`
}

type Header struct {
	Status bool   `json:"status"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type ResponseFull struct {
	Status    int        `json:"status"`
	Condition []swr.Step `json:"condition"`
	Name      string     `json:"name"`
	Headers   []Header   `json:"headers"`
	Body      Body       `json:"body"`
}

func FromResponse(response Response) *ResponseFull {
	result, _ := FromResponseWithOptions(response, swr.DefaultUnmarshalOpts())
	return result
}

func FromResponseWithOptions(response Response, opts swr.UnmarshalOpts) (*ResponseFull, []error) {
	steps, errs := swr.UnmarshalWithOptions(response.Condition, opts)
	if len(errs) > 0 {
		return nil, errs
	}

	return &ResponseFull{
		Status:    response.Status,
		Condition: steps,
		Name:      response.Name,
		Headers:   response.Headers,
		Body:      response.Body,
	}, make([]error, 0)
}
