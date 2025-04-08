package repository

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"maps"
)

type ManagerContext struct {
	mu       sync.Mutex
	context  IRepositoryContext
}

func NewManagerContext(context IRepositoryContext) *ManagerContext {
	return &ManagerContext{
		context:    context,
	}
}

func (m *ManagerContext) Find(owner string, id string) (*context.Context, bool) {
	return m.context.Find(id)
}

func (m *ManagerContext) FindByOwner(owner string) (*context.Context, bool) {
	return m.context.FindByOwner(owner)
}

func (m *ManagerContext) FindByCollection(owner, collection string) (*context.Context, bool) {
	return m.context.FindByCollection(owner, collection)
}

func (m *ManagerContext) Insert(owner string, context *context.Context) *context.Context {
	return m.context.Insert(owner, context)
}

func (m *ManagerContext) InsertFromOwner(owner string, context *context.Context) *context.Context {
	return m.context.InsertFromOwner(owner, context)
}

func (m *ManagerContext) InsertFromCollection(owner, collection string, context *context.Context) *context.Context {
	return m.context.InsertFromCollection(owner, collection, context)
}

func (m *ManagerContext) Delete(context *context.Context) *context.Context {
	return m.context.Delete(context)
}

func (m *ManagerContext) ImportMerge(owner string, target, source *dto.DtoContext) *context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for c, cs := range source.Dictionary {
		category, ok := target.Dictionary[c]
		if !ok {
			category = map[string]dto.DtoItemContext{}
		}

		maps.Copy(category, cs)

		target.Dictionary[c] = category
	}

	ctx := dto.ToContext(target)

	return m.context.Insert(target.Owner, ctx)
}
