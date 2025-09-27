package cookie

import (
	"fmt"
	"strings"
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

func SameSiteFromString(value string) (*SameSite, error) {
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
		return nil, fmt.Errorf("unknown same-site value: '%s'", value)
	}
}
