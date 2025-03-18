package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-collections/collection"
)

type DtoContext struct {
	Id         string                               `json:"_id"`
	Status     bool                                 `json:"status"`
	Timestamp  int64                                `json:"timestamp"`
	Dictionary map[string]map[string]DtoItemContext `json:"dictionary"`
	Owner      string                               `json:"owner"`
	Modified   int64                                `json:"modified"`
}

func NewDtoContextDefault() *DtoContext {
	return &DtoContext{}
}

func (c DtoContext) PersistenceId() string {
	return c.Id
}

func ToContext(dto *DtoContext) *context.Context {
	categories := collection.DictionaryEmpty[string, context.DictionaryVariables]()
	for c, vs := range dto.Dictionary {
		category := collection.DictionaryEmpty[string, context.ItemContext]()
		for k, v := range vs {
			category.Put(k, context.ItemContext{
				Status: v.Status,
				Value:  v.Value,
			})
		}
		categories.Put(c, *category)
	}

	return &context.Context{
		Id:         dto.Id,
		Status:     dto.Status,
		Timestamp:  dto.Timestamp,
		Dictionary: *categories,
		Owner:      dto.Owner,
		Modified:   dto.Modified,
	}
}

func FromContext(ctx *context.Context) *DtoContext {
	categories := map[string]map[string]DtoItemContext{}
	for _, p := range ctx.Dictionary.Pairs() {
		category := map[string]DtoItemContext{}
		values := p.Value()
		for _, v := range values.Pairs() {
			category[v.Key()] = DtoItemContext{
				Status: v.Value().Status,
				Value: v.Value().Value,
			}
		}
		categories[p.Key()] = category
	}

	return &DtoContext{
		Id:         ctx.Id,
		Status:     ctx.Status,
		Timestamp:  ctx.Timestamp,
		Dictionary: categories,
		Owner:      ctx.Owner,
		Modified:   ctx.Modified,
	}
}
