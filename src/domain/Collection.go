package domain

import (
	"slices"
	"sort"
)

type Collection struct {
	Id        string          `json:"_id"`
	Name      string          `json:"name"`
	Timestamp int64           `json:"timestamp"`
	Context   string          `json:"context"`
	Nodes     []NodeReference `json:"nodes"`
	Owner     string          `json:"owner"`
	Modified  int64           `json:"modified"`
}

func NewCollectionDefault() *Collection {
	return &Collection{}
}

func NewCollection(owner string) *Collection {
	return &Collection{
		Id:        "",
		Name:      "",
		Timestamp: 0,
		Context:   "",
		Nodes:     make([]NodeReference, 0),
		Owner:     owner,
		Modified:  0,
	}
}

func (c Collection) ExistsRequest(id string) bool {
	for _, v := range c.Nodes {
		if v.Request == id {
			return true
		}
	}
	return false
}

func (c *Collection) TakeRequest(id string) (*NodeReference, bool) {
	for i, v := range c.Nodes {
		if v.Request == id {
			c.Nodes = slices.Delete(c.Nodes, i, i+1)
			return &v, true
		}
	}
	return nil, false
}

func (c *Collection) SortRequests() *Collection {
	sort.Slice(c.Nodes, func(i, j int) bool {
		return c.Nodes[i].Order < c.Nodes[j].Order
	})
	return c
}

func (c Collection) PersistenceId() string {
	return c.Id
}
