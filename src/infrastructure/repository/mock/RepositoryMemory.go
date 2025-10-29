package mock

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, mock_domain.EndPoint]
	file       repository.IFileManager[mock_domain.EndPoint]
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, mock_domain.EndPoint], file repository.IFileManager[mock_domain.EndPoint]) (*RepositoryMemory, error) {
	requests, err := file.Read()
	if err != nil {
		return nil, err
	}

	return &RepositoryMemory{
		collection: impl.Merge(collection.DictionaryFromMap(requests)),
		file:       file,
	}, nil
}

func (r *RepositoryMemory) Find(id string) (*mock_domain.EndPoint, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(id)
}

func (r *RepositoryMemory) FindByRequest(owner string, method domain.HttpMethod, path string) (*mock_domain.EndPoint, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.FindOne(func(s string, ep mock_domain.EndPoint) bool {
		return ep.Owner == owner && ep.Method == method && ep.Path == path
	})
}

func (r *RepositoryMemory) Insert(owner string, endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	r.muMemory.Lock()
	return r.resolve(owner, endPoint)
}

func (r *RepositoryMemory) resolve(owner string, endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	if endPoint.Id != "" {
		return r.insert(owner, endPoint)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, endPoint)
	}

	endPoint.Id = key

	return r.insert(owner, endPoint)
}

func (r *RepositoryMemory) insert(owner string, endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	defer r.muMemory.Unlock()

	endPoint.Owner = owner

	if endPoint.Timestamp == 0 {
		endPoint.Timestamp = time.Now().UnixMilli()
	}

	endPoint.Modified = time.Now().UnixMilli()

	r.collection.Put(endPoint.Id, *endPoint)

	go r.write(r.collection)

	return endPoint
}

func (r *RepositoryMemory) Delete(endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(endPoint.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, mock_domain.EndPoint]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v mock_domain.EndPoint) any {
		return v
	}).Values()

	err := r.file.Write(items)
	if err != nil {
		log.Error(err)
	}
}
