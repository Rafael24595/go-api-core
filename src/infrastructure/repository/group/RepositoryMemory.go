package historic

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, domain.Group]
	file       repository.IFileManager[domain.Group]
}

func NewRepositoryMemory(
	impl collection.IDictionary[string, domain.Group],
	file repository.IFileManager[domain.Group]) *RepositoryMemory {
	return &RepositoryMemory{
		collection: impl,
		file:       file,
	}
}

func InitializeRepositoryMemory(
	impl collection.IDictionary[string, domain.Group],
	file repository.IFileManager[domain.Group]) (*RepositoryMemory, error) {
	groups, err := file.Read()
	if err != nil {
		return nil, err
	}
	return NewRepositoryMemory(
		impl.Merge(collection.DictionaryFromMap(groups)),
		file), nil
}

func (r *RepositoryMemory) Find(id string) (*domain.Group, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(id)
}

func (r *RepositoryMemory) Insert(owner string, group *domain.Group) *domain.Group {
	r.muMemory.Lock()
	return r.resolve(owner, group)
}

func (r *RepositoryMemory) resolve(owner string, group *domain.Group) *domain.Group {
	if group.Id != "" {
		return r.insert(owner, group)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, group)
	}

	group.Id = key

	return r.insert(owner, group)
}

func (r *RepositoryMemory) insert(owner string, group *domain.Group) *domain.Group {
	defer r.muMemory.Unlock()

	group.Owner = owner

	if group.Timestamp == 0 {
		group.Timestamp = time.Now().UnixMilli()
	}

	group.Modified = time.Now().UnixMilli()

	r.collection.Put(group.Id, *group)

	go r.write(r.collection)

	return group
}

func (r *RepositoryMemory) Delete(context *domain.Group) *domain.Group {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(context.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, domain.Group]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v domain.Group) any {
		return v
	}).Values()

	err := r.file.Write(items)
	if err != nil {
		log.Error(err)
	}
}
