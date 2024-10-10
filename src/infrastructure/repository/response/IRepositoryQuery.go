package response

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type IRepositoryQuery interface {
	FindAll() []domain.Response
	Find(key string) (*domain.Response, bool)
	FindOptions(options repository.FilterOptions[domain.Response]) []domain.Response 
	Exists(key string) bool
	insert(response domain.Response) (domain.Response, []any)
	delete(response domain.Response) (domain.Response, []any)
}
