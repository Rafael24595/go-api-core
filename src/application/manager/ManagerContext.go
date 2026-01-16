package manager

import (
	"sync"

	"maps"

	"github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type ManagerContext struct {
	mu      sync.Mutex
	context context.Repository
}

func NewManagerContext(context context.Repository) *ManagerContext {
	return &ManagerContext{
		context: context,
	}
}

func (m *ManagerContext) Find(owner string, id string) (*context.Context, bool) {
	ctx, exists := m.context.Find(id)
	if !exists || ctx.Owner != owner {
		return nil, false
	}
	return ctx, exists
}

func (m *ManagerContext) Insert(owner string, collection *collection.Collection, context *context.Context) *context.Context {
	if m.isNotOwner(owner, context) {
		return nil
	}
	return m.context.Insert(owner, collection.Id, context)
}

func (m *ManagerContext) ImportMerge(owner string, target, source *dto.DtoContext) *context.Context {
	if source.Owner != owner || target.Owner != owner {
		return nil
	}

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
	ctx, _ = m.context.Update(target.Owner, ctx)

	return ctx
}

func (m *ManagerContext) Update(owner string, context *context.Context) (*context.Context, bool) {
	if m.isNotOwner(owner, context) {
		return nil, false
	}
	return m.context.Update(owner, context)
}

func (m *ManagerContext) Delete(owner string, context *context.Context) *context.Context {
	if context.Owner != owner {
		return nil
	}
	return m.context.Delete(context)
}

func (m *ManagerContext) isNotOwner(owner string, ctx *context.Context) bool {
	return !m.isOwner(owner, ctx)
}

func (m *ManagerContext) isOwner(owner string, ctx *context.Context) bool {
	if ctx == nil {
		return false
	}

	if ctx.Id != "" && ctx.Owner != owner {
		return false
	}

	return true
}
