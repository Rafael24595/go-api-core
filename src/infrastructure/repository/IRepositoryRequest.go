package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryRequest interface {
	Exists(key string) bool
	Find(key string) (*domain.Request, bool)
	FindOptions(options FilterOptions[domain.Request]) []domain.Request
	FindSteps(steps []domain.Historic) []domain.Request
	FindNodes(steps []domain.NodeReference) []domain.Node
	FindAll() []domain.Request
	Insert(owner string, request *domain.Request) *domain.Request
	Delete(request domain.Request) *domain.Request
	DeleteById(id string) *domain.Request 
	DeleteMany(ids ...string) []domain.Request
	DeleteOptions(options FilterOptions[domain.Request]) []string
}
