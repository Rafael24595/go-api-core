package collection

import (
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	muMemory          sync.RWMutex
	muFile            sync.RWMutex
	collection        collection.IDictionary[string, domain.Collection]
	file              repository.IFileManager[domain.Collection]
}

func NewRepositoryMemory(
	impl collection.IDictionary[string, domain.Collection],
	file repository.IFileManager[domain.Collection]) *RepositoryMemory {
	return &RepositoryMemory{
		collection:        impl,
		file:              file,
	}
}

func InitializeRepositoryMemory(
	impl collection.IDictionary[string, domain.Collection],
	file repository.IFileManager[domain.Collection]) (*RepositoryMemory, error) {
	collections, err := file.Read()
	if err != nil {
		return nil, err
	}
	return NewRepositoryMemory(
		impl.Merge(collection.DictionaryFromMap(collections)),
		file), nil
}

func (r *RepositoryMemory) Find(id string) (*domain.Collection, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(id)
}

func (r *RepositoryMemory) FindByOwner(owner string) []domain.Collection {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.FindOptions(repository.FilterOptions[domain.Collection]{
		Predicate: func(c domain.Collection) bool {
			return c.Owner == owner
		},
	})
}

func (r *RepositoryMemory) FindOptions(options repository.FilterOptions[domain.Collection]) []domain.Collection {
	return r.findOptions(options).Collect()
}

func (r *RepositoryMemory) findOptions(options repository.FilterOptions[domain.Collection]) *collection.Vector[domain.Collection] {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	values := r.collection.ValuesVector()

	if options.Predicate != nil {
		values.Filter(options.Predicate)
	}
	if options.Sort != nil {
		values.Sort(options.Sort)
	}

	from := 0
	if options.From != 0 {
		from = options.From
	}

	to := values.Size()
	if options.To != 0 {
		to = options.To
	}

	return values.Slice(from, to)
}

func (r *RepositoryMemory) Exists(key string) bool {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Exists(key)
}

func (r *RepositoryMemory) Insert(owner string, collection *domain.Collection) *domain.Collection {
	r.muMemory.Lock()
	return r.resolve(owner, collection)
}

func (r *RepositoryMemory) resolve(owner string, collection *domain.Collection) *domain.Collection {
	if collection.Id != "" {
		return r.insert(owner, collection)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, collection)
	}

	collection.Id = key

	return r.insert(owner, collection)
}

func (r *RepositoryMemory) insert(owner string, collection *domain.Collection) *domain.Collection {
	r.muMemory.Unlock()

	collection.Owner = owner

	if collection.Timestamp == 0 {
		collection.Timestamp = time.Now().UnixMilli()
	}

	collection.Modified = time.Now().UnixMilli()

	if collection.Name == "" {
		collection.Name = fmt.Sprintf("%s-%d", collection.Owner, collection.Timestamp)
	}

	r.collection.Put(collection.Id, *collection)
	go r.write(r.collection)
	return collection
}

func (r *RepositoryMemory) PushToCollection(owner string, collection *domain.Collection, request *domain.Request) *domain.Collection {
	r.muMemory.Lock()

	if !collection.ExistsRequest(request.Id) {
		collection.Nodes = append(collection.Nodes, domain.NodeReference{
			Order: len(collection.Nodes),
			Request: request.Id,
		})
	}

	return r.resolve(owner, collection)
}

func (r *RepositoryMemory) Delete(collection *domain.Collection) *domain.Collection {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(collection.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, domain.Collection]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v domain.Collection) any {
		return v
	}).Values()

	err := r.file.Write(items)
	if err != nil {
		println(err.Error())
	}
}
