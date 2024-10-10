package request

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type IRepositoryQuery interface {
	FindAll() []domain.Request
	Find(key string) (*domain.Request, bool)
	FindOptions(options repository.FilterOptions[domain.Request]) []domain.Request 
	Exists(key string) bool
	insert(request domain.Request) (domain.Request, []any)
	delete(request domain.Request) (domain.Request, []any)
}