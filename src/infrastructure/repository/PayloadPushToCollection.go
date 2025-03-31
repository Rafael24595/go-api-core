package repository

import "github.com/Rafael24595/go-api-core/src/domain"

type PayloadPushToCollection struct {
	SourceId    string         `json:"source_id"`
	TargetId    string         `json:"target_id"`
	TargetName  string         `json:"target_name"`
	Request     domain.Request `json:"request"`
	RequestName string         `json:"request_name"`
	Movement    Movement       `json:"move"`
}

type Movement string

const (
	CLONE Movement = "clone"
	MOVE  Movement = "move"
)

func (s Movement) String() string {
	return string(s)
}
