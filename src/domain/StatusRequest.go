package domain

import (
	"fmt"
	"strings"
)

type StatusRequest string

const (
	DRAFT StatusRequest = "draft"
	FINAL StatusRequest = "final"
	GROUP StatusRequest = "group"
)

func (s StatusRequest) String() string {
	return string(s)
}

func StatusRequestFromString(value string) (*StatusRequest, error) {
	switch strings.ToLower(value) {
	case string(FINAL):
		status := FINAL
		return &status, nil
	case string(DRAFT):
		status := DRAFT
		return &status, nil
	case string(GROUP):
		status := GROUP
		return &status, nil
	default:
		return nil, fmt.Errorf("unknown same-site value: '%s'", value)
	}
}
