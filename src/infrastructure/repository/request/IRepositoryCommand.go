package request

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type RepositoryCommand interface {
	Insert(request domain.Request) *domain.Request
	Delete(request domain.Request) *domain.Request
}