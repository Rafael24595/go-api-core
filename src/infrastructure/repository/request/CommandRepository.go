package request

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type CommandRepository interface {
	Insert(request domain.Request) *domain.Request
}