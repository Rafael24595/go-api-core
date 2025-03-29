package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryCollection interface {
	Exists(key string) bool
	FindByOwner(owner string) []domain.Collection
	Insert(owner string, collection *domain.Collection) *domain.Collection
	Delete(collection domain.Collection) *domain.Collection
}
