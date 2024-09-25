package auth

import (
	"net/http"
)

type Type string

const (
	Basic  Type = "BASIC"
	Bearer Type = "BEARER"
)

func (t Type) String() string {
	return string(t)
}

func (t Type) LoadStrategy() func(a Auth, r *http.Request) *http.Request {
	switch t {
	case Basic:
		return applyBasicAuth
	case Bearer:
		return applyBearerAuth
	default:
		return applyVoidAuth
	}
}