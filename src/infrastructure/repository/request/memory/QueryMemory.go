package memory

import (
	"sync"

	"github.com/google/uuid"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
)

type QueryMemory struct {
	mu         sync.RWMutex
	collection *collection.CollectionMap[string, domain.Request]
	path       string
}

func NewQueryMemory() *QueryMemory {
	return &QueryMemory{
		collection: collection.EmptyMap[string, domain.Request](),
		path:       DEFAULT_FILE_PATH,
	}
}

func InitializeQueryMemory() (*QueryMemory, error) {
	instance := NewQueryMemory()
	requests, err := instance.read()
	if err != nil {
		return nil, err
	}
	instance.collection = collection.FromMap(requests)
	return instance, nil
}

func (r *QueryMemory) FindAll() []domain.Request {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Values()
}

func (r *QueryMemory) Find(key string) (*domain.Request, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Find(key)
}

func (r *QueryMemory) Exists(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collection.Exists(key)
}

func (r *QueryMemory) insert(request domain.Request) (domain.Request, []any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.insert(request)
	}
	request.Id = key
	r.collection.Put(key, request)
	return request, r.collection.ValuesInterface()
}

func (r *QueryMemory) read() (map[string]domain.Request, error) {
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