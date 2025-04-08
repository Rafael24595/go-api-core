package repository

import (	
	"github.com/Rafael24595/go-api-core/src/domain/context"
)

type IRepositoryContext interface {
	Find(id string) (*context.Context, bool)
	FindByOwner(owner string) (*context.Context, bool)
	FindByCollection(owner, collection string) (*context.Context, bool)
	Insert(owner string, context *context.Context) *context.Context
	InsertFromOwner(owner string, context *context.Context) *context.Context
	InsertFromCollection(owner, collection string, context *context.Context) *context.Context
	Delete(context *context.Context) *context.Context
}
