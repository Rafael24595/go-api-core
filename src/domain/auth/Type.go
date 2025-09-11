package auth

import (
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
