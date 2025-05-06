package response

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, domain.Response]
	file       repository.IFileManager[domain.Response]
}

func NewRepositoryMemory(impl collection.IDictionary[string, domain.Response], file repository.IFileManager[domain.Response]) *RepositoryMemory {
	return &RepositoryMemory{
		collection: impl,
		file:       file,
	}
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, domain.Response], file repository.IFileManager[domain.Response]) (*RepositoryMemory, error) {
	responses, err := file.Read()
	if err != nil {
		return nil, err
	}
	return NewRepositoryMemory(
		impl.Merge(collection.DictionaryFromMap(responses)),
		file), nil
}

func (r *RepositoryMemory) Find(key string) (*domain.Response, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(key)
}

func (r *RepositoryMemory) FindMany(ids []string) []domain.Response {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	responses := make([]domain.Response, 0)
	for _, v := range ids {
		if response, ok := r.collection.Get(v); ok {
			responses = append(responses, *response)
		}
	}

	return responses
}

func (r *RepositoryMemory) Insert(owner string, response *domain.Response) *domain.Response {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	response.Owner = owner

	if response.Id != "" {
		r.collection.Put(response.Id, *response)
		go r.write(r.collection)
		return response
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.Insert(owner, response)
	}

	response.Id = key
	r.collection.Put(key, *response)

	go r.write(r.collection)

	return response
}

func (r *RepositoryMemory) Delete(response *domain.Response) *domain.Response {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()
	
	cursor, _ := r.collection.Remove(response.Id)
	go r.write(r.collection)
	
	return cursor
}

func (r *RepositoryMemory) DeleteMany(responses ...domain.Response) []domain.Response {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	deleted := make([]domain.Response, 0)
	for _, v := range responses {
		cursor, _ := r.collection.Remove(v.Id)
		deleted = append(deleted, *cursor)
	}

	go r.write(r.collection)

	return deleted
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, domain.Response]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v domain.Response) any {
		return v
	}).Values()

	err := r.file.Write(items)
	if err != nil {
		println(err.Error())
	}
}
