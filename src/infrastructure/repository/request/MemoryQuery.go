package request

import (
	"sync"

	"github.com/google/uuid"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type MemoryQuery struct {
	mu         sync.RWMutex
	collection *collection.CollectionMap[string, domain.Request]
	file       repository.IFileManager[domain.Request]
}

func NewMemoryQuery(file repository.IFileManager[domain.Request]) *MemoryQuery {
	return &MemoryQuery{
		collection: collection.EmptyMap[string, domain.Request](),
		file:       file,
	}
}

func InitializeMemoryQuery(file repository.IFileManager[domain.Request]) (*MemoryQuery, error) {
	instance := NewMemoryQuery(file)
	requests, err := instance.file.Read()
	if err != nil {
		return nil, err
	}
	instance.collection = collection.FromMap(requests)
	return instance, nil
}

func (r *MemoryQuery) fileManager() repository.IFileManager[domain.Request] {
	return r.file
}

func (r *MemoryQuery) FindAll() []domain.Request {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Values()
}

func (r *MemoryQuery) FindOptions(options repository.FilterOptions[domain.Request]) []domain.Request {
	return r.findOptions(options).Collect()
}

func (r *MemoryQuery) findOptions(options repository.FilterOptions[domain.Request]) *collection.CollectionList[domain.Request] {
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

func (r *MemoryQuery) Find(key string) (*domain.Request, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Find(key)
}

func (r *MemoryQuery) Exists(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Exists(key)
}

func (r *MemoryQuery) insert(request domain.Request) (domain.Request, []any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if request.Id != "" {
		r.collection.Put(request.Id, request)
		return request, r.collection.ValuesInterface()
	}
	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.insert(request)
	}
	request.Id = key
	r.collection.Put(key, request)
	return request, r.collection.ValuesInterface()
}

func (r *MemoryQuery) delete(request domain.Request) (domain.Request, []any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cursor := request
	old, _ := r.collection.Remove(request.Id, request)
	if old != nil {
		cursor = *old
	}
	return cursor, r.collection.ValuesInterface()
}

func (r *MemoryQuery) deleteOptions(options repository.FilterOptions[domain.Request], mapper func(domain.Request) string) ([]string, []any) {
	optionsCopy := repository.FilterOptions[domain.Request]{
		Predicate: options.Predicate,
		From:      0,
		To:        0,
		Sort:      options.Sort,
	}

	if optionsCopy.Predicate != nil {
		optionsCopy.Predicate = func(r domain.Request) bool {
			return !optionsCopy.Predicate(r)
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
