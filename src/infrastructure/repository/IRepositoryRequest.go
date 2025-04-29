package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type IRepositoryRequest interface {
	Exists(key string) bool
	Find(key string) (*domain.Request, bool)
	FindNodes(steps []domain.NodeReference) []dto.DtoNodeRequest
	FindRequests(steps []domain.NodeReference) []domain.Request
	FindAll() []domain.Request
	Insert(owner string, request *domain.Request) *domain.Request
	InsertMany(owner string, requests []domain.Request) []domain.Request
	Delete(request *domain.Request) *domain.Request
	DeleteById(id string) *domain.Request 
	DeleteMany(ids ...string) []domain.Request
	DeleteOptions(options FilterOptions[domain.Request]) []string
}
