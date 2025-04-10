package context

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/exception"
)

type Domain string

const (
	USER       Domain = "user"
	COLLECTION Domain = "collection"
)

func (s Domain) String() string {
	return string(s)
}

func DomainFromString(value string) (*Domain, exception.ApiError) {
	switch strings.ToLower(value) {
	case string(USER):
		status := USER
		return &status, nil
	case string(COLLECTION):
		status := COLLECTION
		return &status, nil
	default:
		return nil, exception.ApiErrorFrom(422, fmt.Sprintf("Unknown context domain value: '%s'", value))
	}
}
