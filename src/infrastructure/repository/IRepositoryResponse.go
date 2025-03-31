package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryResponse interface {
	Exists(key string) bool
	Find(key string) (*domain.Response, bool)
	FindOptions(options FilterOptions[domain.Response]) []domain.Response
	FindSteps(steps []domain.Historic) []domain.Response
	FindAll() []domain.Response
	Insert(owner string, response *domain.Response) *domain.Response
	Delete(response *domain.Response) *domain.Response
	DeleteById(id string) *domain.Response 
	DeleteMany(ids ...string) []domain.Response
	DeleteOptions(options FilterOptions[domain.Response]) []string
}
