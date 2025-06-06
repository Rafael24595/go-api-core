package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

const(
	BASIC_PARAM_USER = "username"
	BASIC_PARAM_PASSWORD = "password"
)

func applyBasicAuth(a Auth, r *http.Request) *http.Request {
	user := ""
	if pUser, ok := a.Parameters[BASIC_PARAM_USER]; ok {
		user = pUser
	}
	
	password := ""
	if pPassword, ok := a.Parameters[BASIC_PARAM_PASSWORD]; ok {
		user = pPassword
	}

	token := []byte(fmt.Sprintf("%s:%s", user, password))
	token64 := base64.StdEncoding.EncodeToString(token)

	return applyHeaderAuth("Authorization", "Basic", token64, r)
}