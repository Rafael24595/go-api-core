package request

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type QueryRepository interface {
	FindAll() []domain.Request
	Find(key string) (*domain.Request, bool)
	Exists(key string) bool
}