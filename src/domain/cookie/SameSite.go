package cookie

import (
	"fmt"
	"go-api-core/src/commons"
	"strings"
)

type SameSite int

const (
	Strict SameSite = iota
	Lax
	None
)

func (s SameSite) String() string {
	switch s {
	case Strict:
		return "Strict"
	case Lax:
		return "Lax"
	case None:
		return "None"
	default:
		return "Unknown"
	}
}

func SameSiteFromString(value string) (*SameSite, commons.ApiError) {
	switch strings.ToLower(value) {
	case "strict":
		sameSite := Strict
		return &sameSite, nil
	case "lax":
		sameSite := Lax
		return &sameSite, nil
	case "none":
		sameSite := None
		return &sameSite, nil
	default:
		return nil, commons.ApiErrorFrom(422, fmt.Sprintf("Unknown same-site value: '%s'", value))
	}
}