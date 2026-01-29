package collection

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type Repository interface {
	Find(id string) (*Collection, bool)
	FindNodes(steps []domain.NodeReference) []NodeCollection
	Insert(owner string, collection *Collection) *Collection
	Delete(collection *Collection) *Collection
}
