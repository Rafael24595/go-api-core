package auth

import (
	"net/http"
)

func applyBearerAuth(a Auth, r *http.Request) *http.Request {
	prefix := "Bearer"
	if pPrefix, ok := a.Parameters["prefix"]; ok {
		prefix = pPrefix.Value
	}
	
	token := ""
	if pToken, ok := a.Parameters["token"]; ok {
		token = pToken.Value
	}

	return applyHeaderAuth("Authorization", prefix, token, r)
}