package domain

import (
	"errors"
	"fmt"
	"strings"
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

func StatusFromString(value string) (*Status, error) {
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
		return nil, errors.New(fmt.Sprintf("unknown same-site value: '%s'", value))
	}
}
