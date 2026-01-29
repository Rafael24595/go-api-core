package action

import "github.com/Rafael24595/go-api-core/src/domain"

type RepositoryRequest interface {
	Find(id string) (*Request, bool)
	FindMany(ids ...string) []Request
	FindNodes(references []domain.NodeReference) []NodeRequest
	Insert(owner string, request *Request) *Request
	InsertMany(owner string, requests []Request) []Request
	Delete(request *Request) *Request
	DeleteMany(requests ...Request) []Request
}
