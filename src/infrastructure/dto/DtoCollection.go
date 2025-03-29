package dto

import "github.com/Rafael24595/go-api-core/src/domain"

type DtoCollection struct {
	Id        string        `json:"_id"`
	Name      string        `json:"name"`
	Timestamp int64         `json:"timestamp"`
	Context   DtoContext    `json:"context"`
	Nodes     []domain.Node `json:"nodes"`
	Owner     string        `json:"owner"`
	Modified  int64         `json:"modified"`
}
