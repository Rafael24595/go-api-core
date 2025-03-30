package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryCollection interface {
	Exists(key string) bool
	Find(id string) (*domain.Collection, bool)
	FindByOwner(owner string) []domain.Collection
	Insert(owner string, collection *domain.Collection) *domain.Collection
	PushToCollection(owner string, collection *domain.Collection, request *domain.Request) *domain.Collection
	Delete(collection domain.Collection) *domain.Collection
}
