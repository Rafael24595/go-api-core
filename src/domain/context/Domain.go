package context

import (
	"errors"
	"fmt"
	"strings"
)

type Domain string

const (
	USER       Domain = "user"
	COLLECTION Domain = "collection"
)

func (s Domain) String() string {
	return string(s)
}

func DomainFromString(value string) (*Domain, error) {
	switch strings.ToLower(value) {
	case string(USER):
		status := USER
		return &status, nil
	case string(COLLECTION):
		status := COLLECTION
		return &status, nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown context domain value: '%s'", value))
	}
}
