package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryCollection interface {
	Exists(key string) bool
	Find(key string) (*domain.Collection, bool)
	FindOptions(options FilterOptions[domain.Collection]) []domain.Collection 
	FindAll() []domain.Collection
	Insert(collection domain.Collection) *domain.Collection
	Delete(collection domain.Collection) *domain.Collection
}
