package memory

import (
	"sync"

	"github.com/google/uuid"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
)

type QueryMemory struct {
	mu         sync.RWMutex
	collection collection.CollectionMap[string, domain.Request]
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

func (r *QueryMemory) insert(request domain.Request) []domain.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := uuid.New().String()
	if r.Exists(key) {
		return r.insert(request)
	}
	return r.collection.Values()
}
