package collection

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type IRepositoryCommand interface {
	Insert(request domain.Collection) *domain.Collection
	Delete(request domain.Collection) *domain.Collection
	DeleteOptions(options repository.FilterOptions[domain.Collection]) []string
}