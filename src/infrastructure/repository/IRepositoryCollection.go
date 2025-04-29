package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryCollection interface {
	Exists(key string) bool
	Find(id string) (*domain.Collection, bool)
	FindOneBystatus(owner string, Status domain.StatusCollection) (*domain.Collection, bool)
	FindAllBystatus(owner string, Status domain.StatusCollection) []domain.Collection
	FindCollections(steps []domain.NodeReference) []domain.NodeCollection
	Insert(owner string, collection *domain.Collection) *domain.Collection
	PushToCollection(owner string, collection *domain.Collection, request *domain.Request) (*domain.Collection, *domain.Request)
	Delete(collection *domain.Collection) *domain.Collection
}
