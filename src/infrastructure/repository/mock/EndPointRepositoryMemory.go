package mock

import (
	"slices"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-api-core/src/domain"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type EndPointRepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, mock_domain.EndPoint]
	file       repository.IFileManager[mock_domain.EndPoint]
	close      chan bool
}

func InitializeEndPointRepositoryMemory(impl collection.IDictionary[string, mock_domain.EndPoint], file repository.IFileManager[mock_domain.EndPoint]) (*EndPointRepositoryMemory, error) {
	requests, err := file.Read()
	if err != nil {
		return nil, err
	}

	instance := &EndPointRepositoryMemory{
		collection: impl.Merge(collection.DictionaryFromMap(requests)),
		file:       file,
	}

	go instance.watch()

	return instance, nil
}

func (r *EndPointRepositoryMemory) watch() {
	r.once.Do(func() {
		conf := configuration.Instance()
		if !conf.Snapshot().Enable {
			return
		}

		hub := make(chan system.SystemEvent, 1)
		defer close(hub)

		topics := []topic.TopicAction{
			topic_repository.TOPIC_END_POINT.ActionReload(),
		}

		conf.EventHub.Subcribe(repository.RepositoryListener, hub, topics...)
		defer conf.EventHub.Unsubcribe(repository.RepositoryListener, topics...)

		for {
			select {
			case <-r.close:
				log.Customf(repository.SnapshotCategory, "Watcher stopped: local close signal received.")
				return
			case <-hub:
				if err := r.read(); err != nil {
					log.Custome(repository.SnapshotCategory, err)
					return
				}
			case <-conf.Signal.Done():
				log.Customf(repository.SnapshotCategory, "Watcher stopped: global shutdown signal received.")
				return
			}
		}
	})
}

func (r *EndPointRepositoryMemory) read() error {
	requests, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.DictionaryFromMap(requests)
	return nil
}

func (r *EndPointRepositoryMemory) FindAll(owner string) []mock_domain.EndPoint {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	return r.collection.Filter(func(s string, e mock_domain.EndPoint) bool {
		return e.Owner == owner
	}).Values()
}

func (r *EndPointRepositoryMemory) FindAllLite(owner string) []mock_domain.EndPointLite {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	filtered := r.collection.Filter(func(s string, e mock_domain.EndPoint) bool {
		return e.Owner == owner
	}).ValuesVector()

	return collection.VectorMap(filtered, func(e mock_domain.EndPoint) mock_domain.EndPointLite {
		return *mock_domain.LiteFromEndPoint(&e)
	}).Collect()
}

func (r *EndPointRepositoryMemory) FindMany(ids ...string) []mock_domain.EndPoint {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	return r.collection.Filter(func(s string, e mock_domain.EndPoint) bool {
		return slices.Contains(ids, e.Id)
	}).Values()
}

func (r *EndPointRepositoryMemory) Find(id string) (*mock_domain.EndPoint, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	endpoint, ok := r.collection.Get(id)
	return &endpoint, ok
}

func (r *EndPointRepositoryMemory) FindByRequest(owner string, method domain.HttpMethod, path string) (*mock_domain.EndPoint, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	endpoint, ok := r.collection.
		Filter(func(s string, ep mock_domain.EndPoint) bool {
			return ep.Owner == owner && ep.Method == method && ep.Path == path
		}).
		ValuesVector().
		Sort(func(i, j mock_domain.EndPoint) bool {
			return i.Order < j.Order
		}).
		First()

	return &endpoint, ok
}

func (r *EndPointRepositoryMemory) Insert(endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	return r.resolve(endPoint)
}

func (r *EndPointRepositoryMemory) InsertMany(endPoints ...mock_domain.EndPoint) []mock_domain.EndPoint {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	result := make([]mock_domain.EndPoint, len(endPoints))
	for i, v := range endPoints {
		endPoint := r.resolve(&v)
		result[i] = *endPoint
	}

	return result
}

func (r *EndPointRepositoryMemory) resolve(endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	if endPoint.Id != "" {
		return r.insert(endPoint)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(endPoint)
	}

	endPoint.Id = key
	endPoint.Timestamp = time.Now().UnixMilli()
	endPoint.Order = r.collection.Size()

	for i := range endPoint.Responses {
		endPoint.Responses[i].Timestamp = endPoint.Timestamp
	}

	return r.insert(endPoint)
}

func (r *EndPointRepositoryMemory) insert(endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	if endPoint.Timestamp == 0 {
		endPoint.Timestamp = time.Now().UnixMilli()
	}

	endPoint.Modified = time.Now().UnixMilli()

	r.collection.Put(endPoint.Id, *endPoint)

	go r.write(r.collection)

	return endPoint
}

func (r *EndPointRepositoryMemory) Delete(endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(endPoint.Id)
	go r.write(r.collection)

	return &cursor
}

func (r *EndPointRepositoryMemory) write(snapshot collection.IDictionary[string, mock_domain.EndPoint]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
