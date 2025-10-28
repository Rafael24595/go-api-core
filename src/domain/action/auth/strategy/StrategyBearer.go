package auth_strategy

import (
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
)

const (
	DEFAULT_BEARER_PREFIX = "Bearer"
	BEARER_PARAM_PREFIX   = "prefix"
	BEARER_PARAM_TOKEN    = "token"
)

func BearerAuth(status bool, bearer, token string) *auth.Auth {
	return auth.NewAuth(status, auth.Bearer, map[string]string{
		BEARER_PARAM_PREFIX: bearer,
		BEARER_PARAM_TOKEN:  token,
	})
}

func applyBearerAuth(a auth.Auth, r *action.Request) *action.Request {
	prefix := DEFAULT_BEARER_PREFIX
	if pPrefix, ok := a.Parameters[BEARER_PARAM_PREFIX]; ok {
		prefix = pPrefix
	}

	token := ""
	if pToken, ok := a.Parameters[BEARER_PARAM_TOKEN]; ok {
		token = pToken
	}

	return applyHeaderAuth("Authorization", prefix, token, r)
}
