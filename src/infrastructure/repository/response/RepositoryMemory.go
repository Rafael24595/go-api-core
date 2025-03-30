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

func (r *RepositoryMemory) FindOptions(options repository.FilterOptions[domain.Response]) []domain.Response {
	return r.findOptions(options).Collect()
}

func (r *RepositoryMemory) findOptions(options repository.FilterOptions[domain.Response]) *collection.Vector[domain.Response] {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	values := r.collection.ValuesVector()

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

	return values.Slice(from, to)
}

func (r *RepositoryMemory) FindSteps(steps []domain.Historic) []domain.Response {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	responses := []domain.Response{}
	for _, v := range steps {
		if response, ok := r.collection.Get(v.Id); ok {
			responses = append(responses, *response)
		}
	}

	return responses
}

func (r *RepositoryMemory) FindAll() []domain.Response {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Values()
}

func (r *RepositoryMemory) Exists(key string) bool {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Exists(key)
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

func (r *RepositoryMemory) Delete(response domain.Response) *domain.Response {
	return r.DeleteById(response.Id)
}

func (r *RepositoryMemory) DeleteById(id string) *domain.Response {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) DeleteMany(ids ...string) []domain.Response {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	deleted := make([]domain.Response, 0)
	for _, id := range ids {
		cursor, _ := r.collection.Remove(id)
		deleted = append(deleted, *cursor)
	}

	go r.write(r.collection)

	return deleted
}

func (r *RepositoryMemory) DeleteOptions(options repository.FilterOptions[domain.Response]) []string {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	values := r.collection.ValuesVector()

	if options.Predicate != nil {
		values.Filter(func(r domain.Response) bool {
			return !options.Predicate(r)
		})
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

	if to < from {
		to = from
	}

	result := collection.VectorEmpty[domain.Response]().
		Merge(*values.Slice(0, from)).
		Merge(*values.Slice(to, values.Size()))

	r.collection = collection.DictionaryFromVector(*result, func(r domain.Response) string {
		return r.Id
	})

	go r.write(r.collection)

	return r.collection.Keys()
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
