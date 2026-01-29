package mock

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type RepositoryEndPoint interface {
	FindAll(owner string) []EndPoint
	FindMany(ids ...string) []EndPoint
	Find(id string) (*EndPoint, bool)
	FindByRequest(owner string, method domain.HttpMethod, path string) (*EndPoint, bool)
	Insert(endPoint *EndPoint) *EndPoint
	InsertMany(endPoint ...EndPoint) []EndPoint
	Delete(endPoint *EndPoint) *EndPoint
}
