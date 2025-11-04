package mock

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

var defaultResponse = Response{
	Status: 200,
	Headers: map[string]string{
		"content-type": "plain/text",
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
