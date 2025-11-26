package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/collection"
)

type IRepositoryCollection interface {
	Find(id string) (*collection.Collection, bool)
	FindNodes(steps []domain.NodeReference) []collection.NodeCollection
	Insert(owner string, collection *collection.Collection) *collection.Collection
	Delete(collection *collection.Collection) *collection.Collection
}
