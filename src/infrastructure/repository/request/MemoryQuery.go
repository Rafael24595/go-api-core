package request

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

type MemoryQuery struct {
	mu         sync.RWMutex
	collection *collection.CollectionMap[string, domain.Request]
	path       string
}

func NewMemoryQuery() *MemoryQuery {
	return newMemoryQuery(DEFAULT_FILE_PATH)
}

func newMemoryQuery(path string) *MemoryQuery {
	return &MemoryQuery{
		collection: collection.EmptyMap[string, domain.Request](),
		path:       path,
	}
}

func InitializeMemoryQuery() (*MemoryQuery, error) {
	return initializeMemoryQuery(DEFAULT_FILE_PATH)
}

func InitializeMemoryQueryPath(path string) (*MemoryQuery, error) {
	return initializeMemoryQuery(path)
}

func initializeMemoryQuery(path string) (*MemoryQuery, error) {
	instance := newMemoryQuery(path)
	requests, err := instance.read()
	if err != nil {
		return nil, err
	}
	instance.collection = collection.FromMap(requests)
	return instance, nil
}

func (r *MemoryQuery) filePath() string {
	return r.path
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
		From: 0,
		To: 0,
		Sort: options.Sort,
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

func (r *MemoryQuery) read() (map[string]domain.Request, error) {
	buffer, err := utils.ReadFile(r.path)
	if err != nil {
		return nil, err
	}

	if len(buffer) == 0 {
		return make(map[string]domain.Request), nil
	}

	var requests []domain.Request
	err = json.Unmarshal(buffer, &requests)
	if err != nil {
		return nil, err
	}

	return collection.Mapper(requests, func(r domain.Request) string {
		return r.Id
	}).Collect(), nil
}
