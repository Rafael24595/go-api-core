package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/collection"
)

type IRepositoryCollection interface {
	Find(id string) (*collection.Collection, bool)
	FindOneBystatus(owner string, Status collection.StatusCollection) (*collection.Collection, bool)
	FindAllBystatus(owner string, Status collection.StatusCollection) []collection.Collection
	FindCollections(steps []domain.NodeReference) []collection.NodeCollection
	Insert(owner string, collection *collection.Collection) *collection.Collection
	PushToCollection(owner string, collection *collection.Collection, request *action.Request) (*collection.Collection, *action.Request)
	Delete(collection *collection.Collection) *collection.Collection
}
