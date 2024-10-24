package domain

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons"
)

type Status string

const (
	Historic Status = "HIST"
	Saved    Status = "SAVE"
)

func (s Status) String() string {
	return string(s)
}

func StatusFromString(value string) (*Status, commons.ApiError) {
	switch strings.ToLower(value) {
	case "HIST":
		status := Historic
		return &status, nil
	default:
		return nil, commons.ApiErrorFrom(422, fmt.Sprintf("Unknown same-site value: '%s'", value))
	}
}
