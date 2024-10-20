package collection

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type IRepositoryQuery interface {
	fileManager() repository.IFileManager[domain.Collection]
	FindAll() []domain.Collection
	Find(key string) (*domain.Collection, bool)
	FindOptions(options repository.FilterOptions[domain.Collection]) []domain.Collection 
	Exists(key string) bool

	insert(response domain.Collection) (domain.Collection, []any)
	delete(response domain.Collection) (domain.Collection, []any)
	deleteOptions(options repository.FilterOptions[domain.Collection], mapper func(domain.Collection) string) ([]string, []any)
}
