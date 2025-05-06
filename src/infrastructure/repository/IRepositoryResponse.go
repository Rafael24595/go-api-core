package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryResponse interface {
	Find(key string) (*domain.Response, bool)
	FindMany(ids []string) []domain.Response
	Insert(owner string, response *domain.Response) *domain.Response
	Delete(response *domain.Response) *domain.Response 
	DeleteMany(responses ...domain.Response) []domain.Response
}
