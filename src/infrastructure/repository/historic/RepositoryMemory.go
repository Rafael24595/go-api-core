package historic

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, domain.Historic]
	file       repository.IFileManager[domain.Historic]
}

func NewRepositoryMemory(
		impl collection.IDictionary[string, domain.Historic], 
		file repository.IFileManager[domain.Historic]) *RepositoryMemory {
	return &RepositoryMemory{
		collection: impl,
		file:       file,
	}
}

func InitializeRepositoryMemory(
		impl collection.IDictionary[string, domain.Historic], 
		file repository.IFileManager[domain.Historic]) (*RepositoryMemory, error) {
	requests, err := file.Read()
	if err != nil {
		return nil, err
	}
	return NewRepositoryMemory(
		impl.Merge(collection.DictionaryFromMap(requests)),
		file), nil
}

func (r *RepositoryMemory) FindAll() []domain.Historic {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Values()
}

func (r *RepositoryMemory) FindOptions(options repository.FilterOptions[domain.Historic]) []domain.Historic {
	return r.findOptions(options).Collect()
}

func (r *RepositoryMemory) findOptions(options repository.FilterOptions[domain.Historic]) *collection.Vector[domain.Historic] {
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

func (r *RepositoryMemory) Find(key string) (*domain.Historic, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(key)
}

func (r *RepositoryMemory) Exists(key string) bool {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Exists(key)
}

func (r *RepositoryMemory) Insert(request domain.Historic) *domain.Historic {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Put(request.Id, request)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) Delete(request domain.Historic) *domain.Historic {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(request.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) DeleteOptions(options repository.FilterOptions[domain.Historic]) []string {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	values := r.collection.ValuesVector()

	if options.Predicate != nil {
		values.Filter(func(r domain.Historic) bool {
			return !options.Predicate(r)
		})
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

	if to < from {
		to = from
	}

	result := collection.VectorEmpty[domain.Historic]().
		Merge(*values.Slice(0, from)).
		Merge(*values.Slice(to, values.Size()))

	r.collection = collection.DictionaryFromVector(*result, func(r domain.Historic) string {
		return r.Id
	})

	go r.write(r.collection)

	return r.collection.Keys()
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, domain.Historic]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.VectorEmpty[any]().Append(snapshot).Collect()
	err := r.file.Write(items)
	if err != nil {
		println(err.Error())
	}
}
