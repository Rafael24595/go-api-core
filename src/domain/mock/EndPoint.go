package mock

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-collections/collection"
)

var defaultResponse = Response{
	Status: 200,
	Name: "default",
	Headers: []Header{
		{
			Status: true,
			Key: "content-type",
			Value: "plain/text",
		},
	},
	Body: "Default response",
}

type EndPoint struct {
	Id        string              `json:"id"`
	Timestamp int64               `json:"timestamp"`
	Modified  int64               `json:"modified"`
	Name      string              `json:"name"`
	Method    domain.HttpMethod   `json:"method"`
	Path      string              `json:"path"`
	Responses map[string]Response `json:"responses"`
	Safe      bool                `json:"safe"`
	Owner     string              `json:"owner"`
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

func FromEndPoint(endPoint *EndPoint) EndPointLite {
	keys := collection.DictionaryFromMap(endPoint.Responses).Keys()
	return EndPointLite{
		Id:        endPoint.Id,
		Timestamp: endPoint.Timestamp,
		Modified:  endPoint.Modified,
		Name:      endPoint.Name,
		Method:    endPoint.Method,
		Path:      endPoint.Path,
		Responses: keys,
		Safe:      endPoint.Safe,
		Owner:     endPoint.Owner,
	}
}
