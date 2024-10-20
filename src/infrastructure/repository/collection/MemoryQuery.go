package collection

import (
	"sync"

	"github.com/google/uuid"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type MemoryQuery struct {
	mu         sync.RWMutex
	collection *collection.CollectionMap[string, domain.Collection]
	file       repository.IFileManager[domain.Collection]
}

func NewMemoryQuery(file repository.IFileManager[domain.Collection]) *MemoryQuery {
	return &MemoryQuery{
		collection: collection.EmptyMap[string, domain.Collection](),
		file:       file,
	}
}

func InitializeMemoryQuery(file repository.IFileManager[domain.Collection]) (*MemoryQuery, error) {
	instance := NewMemoryQuery(file)
	collections, err := instance.file.Read()
	if err != nil {
		return nil, err
	}
	instance.collection = collection.FromMap(collections)
	return instance, nil
}

func (r *MemoryQuery) fileManager() repository.IFileManager[domain.Collection] {
	return r.file
}

func (r *MemoryQuery) FindAll() []domain.Collection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Values()
}

func (r *MemoryQuery) FindOptions(options repository.FilterOptions[domain.Collection]) []domain.Collection {
	return r.findOptions(options).Collect()
}

func (r *MemoryQuery) findOptions(options repository.FilterOptions[domain.Collection]) *collection.CollectionList[domain.Collection] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	values := r.collection.ValuesCollection()

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

	values.Slice(from, to)

	return values
}

func (r *MemoryQuery) Find(key string) (*domain.Collection, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Find(key)
}

func (r *MemoryQuery) Exists(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Exists(key)
}

func (r *MemoryQuery) insert(collection domain.Collection) (domain.Collection, []any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if collection.Id != "" {
		r.collection.Put(collection.Id, collection)
		return collection, r.collection.ValuesInterface()
	}
	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.insert(collection)
	}
	collection.Id = key
	r.collection.Put(key, collection)
	return collection, r.collection.ValuesInterface()
}

func (r *MemoryQuery) delete(collection domain.Collection) (domain.Collection, []any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cursor := collection
	old, _ := r.collection.Remove(collection.Id, collection)
	if old != nil {
		cursor = *old
	}
	return cursor, r.collection.ValuesInterface()
}

func (r *MemoryQuery) deleteOptions(options repository.FilterOptions[domain.Collection], mapper func(domain.Collection) string) ([]string, []any) {
	optionsCopy := repository.FilterOptions[domain.Collection]{
		Predicate: options.Predicate,
		From:      0,
		To:        0,
		Sort:      options.Sort,
	}

	if optionsCopy.Predicate != nil {
		optionsCopy.Predicate = func(r domain.Collection) bool {
			return !options.Predicate(r)
		}
	}

	filtered := r.findOptions(optionsCopy)

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := filtered.Clone().Slice(0, options.From).
		Merge(*filtered.Slice(options.To, filtered.Size()))

	r.collection = collection.MapperList(*result, mapper)

	return r.collection.Keys(), r.collection.ValuesInterface()
}
