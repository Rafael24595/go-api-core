package auth_strategy

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/auth"
)

func applyVoidAuth(a auth.Auth, r *domain.Request) *domain.Request {
	return r
}
