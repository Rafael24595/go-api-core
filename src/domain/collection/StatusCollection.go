package collection

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/domain/action"
)

type StatusCollection string

const (
	USER StatusCollection = "user"
	TALE StatusCollection = "tale"
	FREE StatusCollection = "free"
)

func (s StatusCollection) String() string {
	return string(s)
}

func StatusCollectionFromString(value string) (*StatusCollection, error) {
	switch strings.ToLower(value) {
	case string(USER):
		status := USER
		return &status, nil
	case string(TALE):
		status := TALE
		return &status, nil
	case string(FREE):
		status := FREE
		return &status, nil
	default:
		return nil, fmt.Errorf("unknown same-site value: '%s'", value)
	}
}

func StatusCollectionToStatusRequest(status *StatusCollection) *action.StatusRequest {
	switch *status {
	case USER:
		status := action.FINAL
		return &status
	case TALE:
		status := action.DRAFT
		return &status
	default:
		status := action.GROUP
		return &status
	}
}
