package auth_strategy

import (
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
)

func applyVoidAuth(a auth.Auth, r *action.Request) *action.Request {
	return r
}
