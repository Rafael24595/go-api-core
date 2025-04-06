package auth

import (
	"net/http"
)

const (
	DEFAULT_BEARER_PREFIX = "Bearer"
	BEARER_PARAM_PREFIX   = "prefix"
	BEARER_PARAM_TOKEN    = "token"
)

func applyBearerAuth(a Auth, r *http.Request) *http.Request {
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
