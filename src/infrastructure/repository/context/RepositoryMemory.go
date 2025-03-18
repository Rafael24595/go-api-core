package historic

import (
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
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

func (r *RepositoryMemory) FindByOwner(owner string) (*context.Context, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(owner)
}

func (r *RepositoryMemory) FindByCollection(owner, collection string) (*context.Context, bool) {
	return r.FindByOwner(fmt.Sprintf("%s-%s", owner, collection))
}

func (r *RepositoryMemory) Insert(owner string, context *context.Context) *context.Context {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	context.Id = owner
	context.Owner = owner

	if context.Timestamp == 0 {
		context.Timestamp = time.Now().UnixMilli()
	}

	context.Modified = time.Now().UnixMilli()

	r.collection.Put(owner, *context)

	go r.write(r.collection)

	return context
}

func (r *RepositoryMemory) Delete(context context.Context) *context.Context {
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
