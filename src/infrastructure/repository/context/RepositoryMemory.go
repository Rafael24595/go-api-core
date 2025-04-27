package historic

import (
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, context.Context]
	file       repository.IFileManager[dto.DtoContext]
}

func NewRepositoryMemory(
	impl collection.IDictionary[string, context.Context],
	file repository.IFileManager[dto.DtoContext]) *RepositoryMemory {
	return &RepositoryMemory{
		collection: impl,
		file:       file,
	}
}

func InitializeRepositoryMemory(
	impl collection.IDictionary[string, context.Context],
	file repository.IFileManager[dto.DtoContext]) (*RepositoryMemory, error) {
	steps, err := file.Read()
	if err != nil {
		return nil, err
	}

	ctx := collection.DictionaryMap(collection.DictionaryFromMap(steps), func(k string, d dto.DtoContext) context.Context {
		return *dto.ToContext(&d)
	})

	return NewRepositoryMemory(
		impl.Merge(ctx),
		file), nil
}

func (r *RepositoryMemory) Find(id string) (*context.Context, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(id)
}

func (r *RepositoryMemory) Insert(owner, collection string, ctx *context.Context) *context.Context {
	return r.resolve(fmt.Sprintf("%s-%s", owner, collection), ctx)
}

func (r *RepositoryMemory) resolve(owner string, ctx *context.Context) *context.Context {
	if ctx.Id != "" {
		return r.insert(owner, ctx)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, ctx)
	}

	ctx.Id = key

	return r.insert(owner, ctx)
}

func (r *RepositoryMemory) insert(owner string, ctx *context.Context) *context.Context {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	ctx.Owner = owner

	if ctx.Timestamp == 0 {
		ctx.Timestamp = time.Now().UnixMilli()
	}

	ctx.Modified = time.Now().UnixMilli()

	r.collection.Put(ctx.Id, *ctx)

	go r.write(r.collection)

	return ctx
}

func (r *RepositoryMemory) Update(owner string, ctx *context.Context) (*context.Context, bool) {
	if _, exists := r.Find(ctx.Id); !exists {
		return ctx, false
	}
	return r.resolve(owner, ctx), true
}

func (r *RepositoryMemory) Delete(context *context.Context) *context.Context {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(context.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, context.Context]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v context.Context) any {
		return *dto.FromContext(&v)
	}).Values()

	err := r.file.Write(items)
	if err != nil {
		println(err.Error())
	}
}
