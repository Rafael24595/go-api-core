package mock

import "github.com/Rafael24595/go-api-core/src/domain"

type EndPoint struct {
	Id        string              `json:"id"`
	Timestamp int64               `json:"timestamp"`
	Modified  int64               `json:"modified"`
	Name      string              `json:"name"`
	Method    domain.HttpMethod   `json:"method"`
	Path      string              `json:"path"`
	Responses map[string]Response `json:"responses"`
	Safe      string              `json:"safe"`
	Owner     string              `json:"owner"`
}

func (r EndPoint) PersistenceId() string {
	return r.Id
}
