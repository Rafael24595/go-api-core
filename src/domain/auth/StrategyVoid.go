package auth

import (
	"net/http"
)

func applyVoidAuth(a Auth, r *http.Request) *http.Request {
	return r
}