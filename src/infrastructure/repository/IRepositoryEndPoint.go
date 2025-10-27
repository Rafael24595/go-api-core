package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain/mock"
)

type IRepositoryEndPoint interface {
	Find(id string) (*mock.EndPoint, bool)
	Insert(owner string, endPoint *mock.EndPoint) *mock.EndPoint
	Delete(endPoint *mock.EndPoint) *mock.EndPoint
}
