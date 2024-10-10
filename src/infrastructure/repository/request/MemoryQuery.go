package request

import (
	"sync"

	"github.com/google/uuid"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
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
	return initializeMemoryQuery(DEFAULT_FILE_PATH)
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

func (r *MemoryQuery) FindAll() []domain.Request {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Values()
}

func (r *MemoryQuery) FindOptions(options repository.FilterOptions[domain.Request]) []domain.Request {
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
	if options.From == 0 {
		from = options.From
	}

	to := values.Size()
	if options.To == 0 {
		to = options.To
	}

	values.Slice(from, to)

	return values.Collect()
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

func (r *MemoryQuery) read() (map[string]domain.Request, error) {
	buffer, err := readFile(r.path)
	if err != nil {
		return nil, err
	}

	requests := map[string]domain.Request{}

	deserializer, err := csvt_translator.NewDeserialzer(string(buffer))
	if err != nil {
		return nil, err
	}
	iterator := deserializer.Iterate()
	for iterator.Next() {
		request := &domain.Request{}
		_ , err := iterator.Deserialize(request)
		if err != nil {
			return nil, err
		}
		requests[request.Id] = *request
	}

	return requests, nil
}
