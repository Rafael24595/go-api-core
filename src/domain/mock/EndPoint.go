package mock

import "github.com/Rafael24595/go-api-core/src/domain"

type EndPoint struct {
	Id        string
	Timestamp int64
	Modified  int64
	Name      string
	Method    domain.HttpMethod
	Uri       string
	Responses map[string]Response
	Safe      string
	Owner     string
}

func (r EndPoint) PersistenceId() string {
	return r.Id
}
