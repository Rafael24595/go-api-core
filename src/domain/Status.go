package domain

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons"
)

type Status string

const (
	DRAFT Status = "draft"
	FINAL Status = "final"
	GROUP Status = "group"
)

func (s Status) String() string {
	return string(s)
}

func StatusFromString(value string) (*Status, commons.ApiError) {
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
		return nil, commons.ApiErrorFrom(422, fmt.Sprintf("Unknown same-site value: '%s'", value))
	}
}
