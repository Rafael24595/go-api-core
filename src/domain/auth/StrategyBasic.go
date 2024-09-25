package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

func applyBasicAuth(a Auth, r *http.Request) *http.Request {
	user := ""
	if pUser, ok := a.Parameters["username"]; ok {
		user = pUser.Value
	}
	
	password := ""
	if pPassword, ok := a.Parameters["password"]; ok {
		user = pPassword.Value
	}

	token := []byte(fmt.Sprintf("%s:%s", user, password))
	token64 := base64.StdEncoding.EncodeToString(token)

	return applyHeaderAuth("Authorization", "Basic", token64, r)
}