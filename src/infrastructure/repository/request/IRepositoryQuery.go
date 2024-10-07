package request

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type RepositoryQuery interface {
	FindAll() []domain.Request
	Find(key string) (*domain.Request, bool)
	Exists(key string) bool
	insert(request domain.Request) (domain.Request, []any)
	delete(request domain.Request) (domain.Request, []any)
}
