package auth

import (
	"net/http"
	"strings"
)

type Type string

const (
	None   Type = "NONE"
	Basic  Type = "BASIC"
	Bearer Type = "BEARER"
)

func TypeFromString(typ string) (Type, bool) {
	switch strings.ToUpper(typ) {
	case Basic.String():
		return Basic, true
	case Bearer.String():
		return Bearer, true
	default:
		return None, false
	}
}

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
