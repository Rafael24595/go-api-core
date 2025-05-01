package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type IRepositoryRequest interface {
	Find(key string) (*domain.Request, bool)
	FindMany(ids []string) []domain.Request
	FindNodes(steps []domain.NodeReference) []dto.DtoNodeRequest
	FindRequests(steps []domain.NodeReference) []domain.Request
	Insert(owner string, request *domain.Request) *domain.Request
	InsertMany(owner string, requests []domain.Request) []domain.Request
	Delete(request *domain.Request) *domain.Request 
	DeleteMany(requests ...domain.Request) []domain.Request
}
