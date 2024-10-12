package request

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type IRepositoryCommand interface {
	Insert(request domain.Request) *domain.Request
	Delete(request domain.Request) *domain.Request
	DeleteOptions(options repository.FilterOptions[domain.Request]) []string
}