package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryHistoric interface {
	Exists(key string) bool
	Find(key string) (*domain.Historic, bool)
	FindByOwner(owner string) []domain.Historic
	FindAll() []domain.Historic
	Insert(request domain.Historic) *domain.Historic
	Delete(request domain.Historic) *domain.Historic
	DeleteOptions(options FilterOptions[domain.Historic]) []string
}
