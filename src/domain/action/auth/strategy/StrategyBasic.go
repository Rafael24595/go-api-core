package auth_strategy

import (
	"encoding/base64"
	"fmt"

	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
)

const (
	BASIC_PARAM_USER     = "username"
	BASIC_PARAM_PASSWORD = "password"
)

func BasicAuth(status bool, user, pass string) *auth.Auth {
	return auth.NewAuth(status, auth.Basic, map[string]string{
		BASIC_PARAM_USER:     user,
		BASIC_PARAM_PASSWORD: pass,
	})
}

func applyBasicAuth(a auth.Auth, r *action.Request) *action.Request {
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
