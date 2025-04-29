package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
)

type IRepositoryGroup interface {
	Exists(key string) bool
	Find(id string) (*domain.Group, bool)
	Insert(owner string, group *domain.Group) *domain.Group
	Delete(group *domain.Group) *domain.Group
}
