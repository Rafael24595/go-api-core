package request

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
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

func (r *RepositoryMemory) FindOptions(options repository.FilterOptions[domain.Request]) []domain.Request {
	return r.findOptions(options).Collect()
}

func (r *RepositoryMemory) findOptions(options repository.FilterOptions[domain.Request]) *collection.Vector[domain.Request] {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	values := r.collection.ValuesVector()

	if options.Predicate != nil {
		values.FilterSelf(options.Predicate)
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

func (r *RepositoryMemory) FindSteps(steps []domain.Historic) []domain.Request {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := []domain.Request{}
	for _, v := range steps {
		if request, ok := r.collection.Get(v.Id); ok {
			requests = append(requests, *request)
		}
	}

	return requests
}

func (r *RepositoryMemory) FindAll() []domain.Request {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Values()
}

func (r *RepositoryMemory) Exists(key string) bool {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Exists(key)
}

func (r *RepositoryMemory) Insert(owner string, request *domain.Request) *domain.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	request.Owner = owner

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().UnixMilli()
	}

	request.Modified = time.Now().UnixMilli()

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

func (r *RepositoryMemory) Delete(request domain.Request) *domain.Request {
	return r.DeleteById(request.Id)
}

func (r *RepositoryMemory) DeleteById(id string) *domain.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) DeleteOptions(options repository.FilterOptions[domain.Request]) []string {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	values := r.collection.ValuesVector()

	if options.Predicate != nil {
		values.Filter(func(r domain.Request) bool {
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

	result := collection.VectorEmpty[domain.Request]().
		Merge(*values.Slice(0, from)).
		Merge(*values.Slice(to, values.Size()))

	r.collection = collection.DictionaryFromVector(*result, func(r domain.Request) string {
		return r.Id
	})

	go r.write(r.collection)

	return r.collection.Keys()
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
