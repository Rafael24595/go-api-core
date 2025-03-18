package repository

import (	
	"github.com/Rafael24595/go-api-core/src/domain/context"
)

type IRepositoryContext interface {
	FindByOwner(owner string) (*context.Context, bool)
	FindByCollection(owner, collection string) (*context.Context, bool)
	Insert(owner string, context *context.Context) *context.Context
	Delete(context context.Context) *context.Context
}
