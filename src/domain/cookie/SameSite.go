package cookie

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/exception"
)

type SameSite string

const (
	Strict SameSite = "Strict"
	Lax    SameSite = "Lax"
	None   SameSite = "None"
)

func (s SameSite) String() string {
	return string(s)
}

func SameSiteFromString(value string) (*SameSite, exception.ApiError) {
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
		return nil, exception.ApiErrorFrom(422, fmt.Sprintf("Unknown same-site value: '%s'", value))
	}
}
