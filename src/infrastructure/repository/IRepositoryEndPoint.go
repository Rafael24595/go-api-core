package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/mock"
)

type IRepositoryEndPoint interface {
	FindAll(owner string) []mock.EndPoint
	FindMany(ids ...string) []mock.EndPoint
	Find(id string) (*mock.EndPoint, bool)
	FindByRequest(owner string, method domain.HttpMethod, path string) (*mock.EndPoint, bool)
	Insert(endPoint *mock.EndPoint) *mock.EndPoint
	InsertMany(endPoint ...mock.EndPoint) []mock.EndPoint
	Delete(endPoint *mock.EndPoint) *mock.EndPoint
}
