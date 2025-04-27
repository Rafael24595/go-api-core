package repository

import (	
	"github.com/Rafael24595/go-api-core/src/domain/context"
)

type IRepositoryContext interface {
	Find(id string) (*context.Context, bool)
	Insert(owner string, collection string, context *context.Context) *context.Context
	Update(owner string, context *context.Context) (*context.Context, bool)
	Delete(context *context.Context) *context.Context
}
