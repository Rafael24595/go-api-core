package mock

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	"github.com/Rafael24595/go-collections/collection"
)

var defaultResponse = Response{
	Status: 200,
	Name:   "default",
	Headers: []Header{
		{
			Status: true,
			Key:    "content-type",
			Value:  "plain/text",
		},
	},
	Body: Body{
		ContentType: domain.Text,
		Payload: "Default response",
	},
}

type EndPoint struct {
	Id        string            `json:"id"`
	Timestamp int64             `json:"timestamp"`
	Modified  int64             `json:"modified"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Path      string            `json:"path"`
	Responses []Response        `json:"responses"`
	Safe      bool              `json:"safe"`
	Owner     string            `json:"owner"`
}

func (r EndPoint) DefaultResponse() Response {
	return defaultResponse
}

func (r EndPoint) PersistenceId() string {
	return r.Id
}

type EndPointLite struct {
	Id        string            `json:"id"`
	Timestamp int64             `json:"timestamp"`
	Modified  int64             `json:"modified"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Path      string            `json:"path"`
	Responses []string          `json:"responses"`
	Safe      bool              `json:"safe"`
	Owner     string            `json:"owner"`
}

func LiteFromEndPoint(endPoint *EndPoint)* EndPointLite {
	keys := collection.MapToVector(endPoint.Responses, func(r Response) string {
		return r.Name
	}, collection.MakeVector)

	return &EndPointLite{
		Id:        endPoint.Id,
		Timestamp: endPoint.Timestamp,
		Modified:  endPoint.Modified,
		Name:      endPoint.Name,
		Method:    endPoint.Method,
		Path:      endPoint.Path,
		Responses: keys.Collect(),
		Safe:      endPoint.Safe,
		Owner:     endPoint.Owner,
	}
}

type EndPointFull struct {
	Id        string            `json:"id"`
	Timestamp int64             `json:"timestamp"`
	Modified  int64             `json:"modified"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Path      string            `json:"path"`
	Responses []ResponseFull    `json:"responses"`
	Safe      bool              `json:"safe"`
	Owner     string            `json:"owner"`
}

func FullFromEndPoint(endPoint *EndPoint) (*EndPointFull, []error) {
	responses := make([]ResponseFull, len(endPoint.Responses))

	opts := swr.UnmarshalOpts{ Evalue: true }
	for i, v := range endPoint.Responses {
		result, errs := FromResponseWithOptions(v, opts)
		if len(errs) > 0 {
			return nil, errs
		}

		responses[i] = *result
	}

	return &EndPointFull{
		Id:        endPoint.Id,
		Timestamp: endPoint.Timestamp,
		Modified:  endPoint.Modified,
		Name:      endPoint.Name,
		Method:    endPoint.Method,
		Path:      endPoint.Path,
		Responses: responses,
		Safe:      endPoint.Safe,
		Owner:     endPoint.Owner,
	}, make([]error, 0)
}

func ToEndPointFromFull(endPoint *EndPointFull) (*EndPoint, []error) {
	responses := make([]Response, len(endPoint.Responses))

	opts := swr.MarshalOpts{ Evalue: true }
	for i, v := range endPoint.Responses {
		result, errs := ToResponseWithOptions(v, opts)
		if len(errs) > 0 {
			return nil, errs
		}

		responses[i] = *result
	}

	return &EndPoint{
		Id:        endPoint.Id,
		Timestamp: endPoint.Timestamp,
		Modified:  endPoint.Modified,
		Name:      endPoint.Name,
		Method:    endPoint.Method,
		Path:      endPoint.Path,
		Responses: responses,
		Safe:      endPoint.Safe,
		Owner:     endPoint.Owner,
	}, make([]error, 0)
}