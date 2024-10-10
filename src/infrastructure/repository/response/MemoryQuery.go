package response

import (
	"sync"

	"github.com/google/uuid"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

type MemoryQuery struct {
	mu         sync.RWMutex
	collection *collection.CollectionMap[string, domain.Response]
	path       string
}

func NewMemoryQuery() *MemoryQuery {
	return newMemoryQuery(DEFAULT_FILE_PATH)
}

func newMemoryQuery(path string) *MemoryQuery {
	return &MemoryQuery{
		collection: collection.EmptyMap[string, domain.Response](),
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
	responses, err := instance.read()
	if err != nil {
		return nil, err
	}
	instance.collection = collection.FromMap(responses)
	return instance, nil
}

func (r *MemoryQuery) FindAll() []domain.Response {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Values()
}

func (r *MemoryQuery) FindOptions(options repository.FilterOptions[domain.Response]) []domain.Response {
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

func (r *MemoryQuery) Find(key string) (*domain.Response, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Find(key)
}

func (r *MemoryQuery) Exists(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Exists(key)
}

func (r *MemoryQuery) insert(response domain.Response) (domain.Response, []any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if response.Id != "" {
		r.collection.Put(response.Id, response)
		return response, r.collection.ValuesInterface()
	}
	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.insert(response)
	}
	response.Id = key
	r.collection.Put(key, response)
	return response, r.collection.ValuesInterface()
}

func (r *MemoryQuery) delete(response domain.Response) (domain.Response, []any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cursor := response
	old, _ := r.collection.Remove(response.Id, response)
	if old != nil {
		cursor = *old
	}
	return cursor, r.collection.ValuesInterface()
}

func (r *MemoryQuery) read() (map[string]domain.Response, error) {
	buffer, err := utils.ReadFile(r.path)
	if err != nil {
		return nil, err
	}

	responses := map[string]domain.Response{}

	deserializer, err := csvt_translator.NewDeserialzer(string(buffer))
	if err != nil {
		return nil, err
	}
	iterator := deserializer.Iterate()
	for iterator.Next() {
		response := &domain.Response{}
		_ , err := iterator.Deserialize(response)
		if err != nil {
			return nil, err
		}
		responses[response.Id] = *response
	}

	return responses, nil
}
