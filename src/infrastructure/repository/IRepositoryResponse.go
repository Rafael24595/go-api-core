package repository

import "github.com/Rafael24595/go-api-core/src/domain/action"

type IRepositoryResponse interface {
	Find(key string) (*action.Response, bool)
	FindMany(ids []string) []action.Response
	Insert(owner string, response *action.Response) *action.Response
	Delete(response *action.Response) *action.Response
	DeleteMany(responses ...action.Response) []action.Response
}
