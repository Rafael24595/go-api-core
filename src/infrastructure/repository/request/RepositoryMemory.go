package request

import (
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, domain.Request]
	file       repository.IFileManager[domain.Request]
}

func NewRepositoryMemory(impl collection.IDictionary[string, domain.Request], file repository.IFileManager[domain.Request]) *RepositoryMemory {
	return &RepositoryMemory{
		collection: impl,
		file:       file,
	}
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, domain.Request], file repository.IFileManager[domain.Request]) (*RepositoryMemory, error) {
	requests, err := file.Read()
	if err != nil {
		return nil, err
	}
	return NewRepositoryMemory(
		impl.Merge(collection.DictionaryFromMap(requests)),
		file), nil
}

func (r *RepositoryMemory) Find(key string) (*domain.Request, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(key)
}

func (r *RepositoryMemory) FindMany(ids []string) []domain.Request {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := make([]domain.Request, 0)
	for _, v := range ids {
		if request, ok := r.collection.Get(v); ok {
			requests = append(requests, *request)
		}
	}

	return requests
}

func (r *RepositoryMemory) FindNodes(references []domain.NodeReference) []dto.DtoNodeRequest {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := make([]dto.DtoNodeRequest, 0)
	for _, v := range references {
		if request, ok := r.collection.Get(v.Item); ok {
			requests = append(requests, dto.DtoNodeRequest{
				Order:   v.Order,
				Request: *dto.FromRequest(request),
			})
		}
	}

	return requests
}

func (r *RepositoryMemory) FindRequests(references []domain.NodeReference) []domain.Request {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := make([]domain.Request, 0)
	for _, v := range references {
		if request, ok := r.collection.Get(v.Item); ok {
			requests = append(requests, *request)
		}
	}

	return requests
}

func (r *RepositoryMemory) Insert(owner string, request *domain.Request) *domain.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	request.Owner = owner

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().UnixMilli()
	}

	request.Modified = time.Now().UnixMilli()

	if request.Name == "" {
		request.Name = fmt.Sprintf("%s-%s-%d", request.Owner, request.Method, request.Timestamp)
	}

	if request.Id != "" {
		r.collection.Put(request.Id, *request)
		go r.write(r.collection)
		return request
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.Insert(owner, request)
	}

	request.Id = key
	r.collection.Put(key, *request)

	go r.write(r.collection)

	return request
}

func (r *RepositoryMemory) InsertMany(owner string, requests []domain.Request) []domain.Request {
	for i, v := range requests {
		req := r.Insert(owner, &v)
		requests[i] = *req
	}
	return requests
}

func (r *RepositoryMemory) Delete(request *domain.Request) *domain.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(request.Id)
	if cursor != nil {
		go r.write(r.collection)
	}

	return cursor
}

func (r *RepositoryMemory) DeleteMany(requests ...domain.Request) []domain.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	deleted := make([]domain.Request, 0)
	for _, v := range requests {
		cursor, _ := r.collection.Remove(v.Id)
		deleted = append(deleted, *cursor)
	}

	go r.write(r.collection)

	return deleted
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, domain.Request]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v domain.Request) any {
		return v
	}).Values()

	err := r.file.Write(items)
	if err != nil {
		println(err.Error())
	}
}
