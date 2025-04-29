package domain

import (
	"slices"
	"sort"
)

type Group struct {
	Id        string          `json:"_id"`
	Timestamp int64           `json:"timestamp"`
	Nodes     []NodeReference `json:"nodes"`
	Owner     string          `json:"owner"`
	Modified  int64           `json:"modified"`
}

func NewGroupDefault() *Group {
	return &Group{}
}

func NewGroup(owner string) *Group {
	return &Group{
		Id:        "",
		Timestamp: 0,
		Nodes:     make([]NodeReference, 0),
		Owner:     owner,
		Modified:  0,
	}
}

func (c Group) ExistsNode(id string) bool {
	for _, v := range c.Nodes {
		if v.Item == id {
			return true
		}
	}
	return false
}

func (c *Group) TakeNode(id string) (*NodeReference, bool) {
	for i, v := range c.Nodes {
		if v.Item == id {
			c.Nodes = slices.Delete(c.Nodes, i, i+1)
			return &v, true
		}
	}
	return nil, false
}

func (c *Group) ResolveNode(node *NodeReference) (*NodeReference, bool) {
	for i, v := range c.Nodes {
		if v.Item == node.Item {
			c.Nodes[i] = *node
			return &v, true
		}
	}
	c.Nodes = append(c.Nodes, *node)
	return nil, false
}

func (c *Group) SortNodes() *Group {
	sort.Slice(c.Nodes, func(i, j int) bool {
		return c.Nodes[i].Order < c.Nodes[j].Order
	})
	return c
}

func (c *Group) FixNodesOrder() *Group {
	for i := range c.Nodes {
		c.Nodes[i].Order = i
	}
	return c
}

func (c Group) PersistenceId() string {
	return c.Id
}
