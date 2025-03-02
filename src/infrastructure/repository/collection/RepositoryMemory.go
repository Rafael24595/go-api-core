package collection

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, domain.Collection]
	file       repository.IFileManager[domain.Collection]
}

func NewRepositoryMemory(impl collection.IDictionary[string, domain.Collection], file repository.IFileManager[domain.Collection]) *RepositoryMemory {
	return &RepositoryMemory{
		collection: impl,
		file:       file,
	}
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, domain.Collection], file repository.IFileManager[domain.Collection]) (*RepositoryMemory, error) {
	collections, err := file.Read()
	if err != nil {
		return nil, err
	}
	return NewRepositoryMemory(
		impl.Merge(collection.DictionaryFromMap(collections)),
		file), nil
}

func (r *RepositoryMemory) FindAll() []domain.Collection {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Values()
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

func (r *RepositoryMemory) Find(key string) (*domain.Collection, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(key)
}

func (r *RepositoryMemory) Exists(key string) bool {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Exists(key)
}

func (r *RepositoryMemory) Insert(collection domain.Collection) *domain.Collection {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Put(collection.Id, collection)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) Delete(collection domain.Collection) *domain.Collection {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(collection.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, domain.Collection]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.VectorEmpty[any]().Append(snapshot).Collect()
	err := r.file.Write(items)
	if err != nil {
		println(err.Error())
	}
}
