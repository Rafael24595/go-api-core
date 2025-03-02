package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryResponse interface {
	Exists(key string) bool
	Find(key string) (*domain.Response, bool)
	FindOptions(options FilterOptions[domain.Response]) []domain.Response
	FindAll() []domain.Response
	Insert(response domain.Response) *domain.Response
	Delete(response domain.Response) *domain.Response
	DeleteOptions(options FilterOptions[domain.Response]) []string
}
