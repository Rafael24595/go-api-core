package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
)

type IRepositoryRequest interface {
	Find(id string) (*action.Request, bool)
	FindMany(ids ...string) []action.Request
	FindNodes(references []domain.NodeReference) []action.NodeRequest
	Insert(owner string, request *action.Request) *action.Request
	InsertMany(owner string, requests []action.Request) []action.Request
	Delete(request *action.Request) *action.Request
	DeleteMany(requests ...action.Request) []action.Request
}
