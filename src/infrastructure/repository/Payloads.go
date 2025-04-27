package repository

import (
	"sort"

	"github.com/Rafael24595/go-api-core/src/domain"
)

type Movement string

const (
	CLONE Movement = "clone"
	MOVE  Movement = "move"
)

func (s Movement) String() string {
	return string(s)
}

type PayloadCollectRequest struct {
	SourceId    string         `json:"source_id"`
	TargetId    string         `json:"target_id"`
	TargetName  string         `json:"target_name"`
	Request     domain.Request `json:"request"`
	RequestName string         `json:"request_name"`
	Movement    Movement       `json:"move"`
}

type PayloadSortCollection struct {
	Nodes []PayloadCollectionNode `json:"nodes"`
}

func (p *PayloadSortCollection) SortRequests() *PayloadSortCollection {
	sort.Slice(p.Nodes, func(i, j int) bool {
		return p.Nodes[i].Order < p.Nodes[j].Order
	})
	return p
}

type PayloadCollectionNode struct {
	Order   int    `json:"order"`
	Request string `json:"request"`
}
