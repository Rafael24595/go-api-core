package response

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type IRepositoryCommand interface {
	Insert(request domain.Response) *domain.Response
	Delete(request domain.Response) *domain.Response
	DeleteOptions(options repository.FilterOptions[domain.Response]) []string
}