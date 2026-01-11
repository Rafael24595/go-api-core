package mock

import (
	"fmt"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	"github.com/Rafael24595/go-collections/collection"
)

const DefaultResponse = "default"

func defaultResponse() *Response {
	return &Response{
		Order:  0,
		Status: true,
		Code:   200,
		Name:   DefaultResponse,
		Arguments: []Argument{
			{
				Status: true,
				Key:    "content-type",
				Value:  "plain/text",
			},
		},
		Body: Body{
			ContentType: domain.Text,
			Payload:     "Default response",
		},
	}
}

type Response struct {
	Order     int        `json:"order"`
	Status    bool       `json:"status"`
	Code      int        `json:"code"`
	Timestamp int64      `json:"timestamp"`
	Condition string     `json:"condition"`
	Name      string     `json:"name"`
	Arguments []Argument `json:"arguments"`
	Body      Body       `json:"body"`
}

type Body struct {
	ContentType domain.ContentType `json:"content_type"`
	Payload     string             `json:"payload"`
}

type Argument struct {
	Status bool   `json:"status"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type ResponseFull struct {
	Order     int        `json:"order"`
	Status    bool       `json:"status"`
	Code      int        `json:"code"`
	Timestamp int64      `json:"timestamp"`
	Condition []swr.Step `json:"condition"`
	Name      string     `json:"name"`
	Arguments []Argument `json:"arguments"`
	Body      Body       `json:"body"`
}

func FixResponses(responses []Response) []Response {
	var defRes *Response

	coll := collection.VectorFromList(responses)

	index := coll.IndexOf(func(r Response) bool { return r.Name == DefaultResponse })
	if index != -1 {
		def, _ := coll.Remove(index)
		def.Condition = ""
		defRes = def
	}

	if defRes == nil {
		defRes = defaultResponse()
	}

	defRes.Status = true
	defRes.Order = 0

	coll.FilterSelf(func(r Response) bool { return r.Name != DefaultResponse })

	coll.Sort(func(i, j Response) bool {
		return i.Order < j.Order
	})

	coll.Unshift(*defRes)

	coll.Map(func(i int, r Response) Response {
		r.Order = i
		return r
	})

	return fixResponsesData(coll.Collect())
}

func fixResponsesData(responses []Response) []Response {
	cache := make(map[string]bool, 0)

	for i, v := range responses {
		name := fixResponseName(v, cache)
		responses[i].Name = name
		cache[name] = true

		if responses[i].Timestamp == 0 {
			responses[i].Timestamp = time.Now().UnixMilli()
		}
	}

	return responses
}

func fixResponseName(response Response, cache map[string]bool) string {
	name := response.Name
	count := 1

	for cache[name] {
		name = fmt.Sprintf("%s-copy-%d", response.Name, count)
		count++
	}

	return name
}

func FromResponse(response Response) *ResponseFull {
	result, _ := FromResponseWithOptions(response, swr.DefaultUnmarshalOpts())
	return result
}

func FromResponseWithOptions(response Response, opts swr.UnmarshalOpts) (*ResponseFull, []error) {
	steps, errs := swr.UnmarshalWithOptions(response.Condition, opts)
	return &ResponseFull{
		Order:     response.Order,
		Status:    response.Status,
		Code:      response.Code,
		Timestamp: response.Timestamp,
		Condition: steps,
		Name:      response.Name,
		Arguments: response.Arguments,
		Body:      response.Body,
	}, errs
}

func ToResponse(response ResponseFull) *Response {
	result, _ := ToResponseWithOptions(response, swr.DefaultMarshalOpts())
	return result
}

func ToResponseWithOptions(response ResponseFull, opts swr.MarshalOpts) (*Response, []error) {
	condition, errs := swr.MarshalWithOptions(response.Condition, opts)
	return &Response{
		Order:     response.Order,
		Status:    response.Status,
		Code:      response.Code,
		Timestamp: response.Timestamp,
		Condition: condition,
		Name:      response.Name,
		Arguments: response.Arguments,
		Body:      response.Body,
	}, errs
}
