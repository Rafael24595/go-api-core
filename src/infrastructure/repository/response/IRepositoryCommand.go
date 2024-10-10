package response

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryCommand interface {
	Insert(request domain.Response) *domain.Response
	Delete(request domain.Response) *domain.Response
}