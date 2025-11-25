package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type IRepositoryRequest interface {
	Find(key string) (*action.Request, bool)
	FindMany(ids ...string) []action.Request
	FindLiteNodes(steps []domain.NodeReference) []dto.DtoLiteNodeRequest
	FindNodes(steps []domain.NodeReference) []dto.DtoNodeRequest
	Insert(owner string, request *action.Request) *action.Request
	InsertMany(owner string, requests []action.Request) []action.Request
	Delete(request *action.Request) *action.Request
	DeleteMany(requests ...action.Request) []action.Request
}
