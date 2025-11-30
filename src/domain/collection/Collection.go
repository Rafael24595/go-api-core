package collection

import (
	"slices"
	"sort"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
)

type Collection struct {
	Id        string                 `json:"_id"`
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Context   string                 `json:"context"`
	Nodes     []domain.NodeReference `json:"nodes"`
	Owner     string                 `json:"owner"`
	Modified  int64                  `json:"modified"`
	Status    StatusCollection       `json:"status"`
}

func NewUserCollection(owner string) *Collection {
	return newCollection(owner, USER)
}

func NewTaleCollection(owner string) *Collection {
	return newCollection(owner, TALE)
}

func NewFreeCollection(owner string) *Collection {
	return newCollection(owner, FREE)
}

func newCollection(owner string, status StatusCollection) *Collection {
	return &Collection{
		Id:        "",
		Name:      "",
		Timestamp: 0,
		Context:   "",
		Nodes:     make([]domain.NodeReference, 0),
		Owner:     owner,
		Modified:  0,
		Status:    status,
	}
}

func (c Collection) ExistsRequest(id string) bool {
	for _, v := range c.Nodes {
		if v.Item == id {
			return true
		}
	}
	return false
}

func (c *Collection) TakeRequest(id string) (*domain.NodeReference, bool) {
	for i, v := range c.Nodes {
		if v.Item == id {
			c.Nodes = slices.Delete(c.Nodes, i, i+1)
			return &v, true
		}
	}
	return nil, false
}

func (c *Collection) ResolveRequest(item string) (*domain.NodeReference, bool) {
	for i, v := range c.Nodes {
		if v.Item == item {
			c.Nodes[i] = domain.NodeReference{
				Order: v.Order,
				Item:  item,
			}
			return &v, true
		}
	}
	c.Nodes = append(c.Nodes, domain.NodeReference{
		Order: len(c.Nodes),
		Item:  item,
	})
	return nil, false
}

func (c *Collection) SortRequests() *Collection {
	sort.Slice(c.Nodes, func(i, j int) bool {
		return c.Nodes[i].Order < c.Nodes[j].Order
	})
	return c
}

func (c *Collection) FixRequestsOrder() *Collection {
	for i := range c.Nodes {
		c.Nodes[i].Order = i
	}
	return c
}

func (c Collection) PersistenceId() string {
	return c.Id
}

type CollectionLite struct {
	Id        string                   `json:"_id"`
	Name      string                   `json:"name"`
	Timestamp int64                    `json:"timestamp"`
	Context   string                   `json:"context"`
	Nodes     []action.NodeRequestLite `json:"nodes"`
	Owner     string                   `json:"owner"`
	Modified  int64                    `json:"modified"`
	Status    StatusCollection         `json:"status"`
}

func ToLiteCollection(collection *Collection, ctx string, nodes []action.NodeRequestLite) *CollectionLite {
	return &CollectionLite{
		Id:        collection.Id,
		Name:      collection.Name,
		Timestamp: collection.Timestamp,
		Context:   ctx,
		Nodes:     nodes,
		Owner:     collection.Owner,
		Modified:  collection.Modified,
	}
}
