package mock

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/domain"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, mock_domain.EndPoint]
	file       repository.IFileManager[mock_domain.EndPoint]
	close      chan bool
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, mock_domain.EndPoint], file repository.IFileManager[mock_domain.EndPoint]) (*RepositoryMemory, error) {
	requests, err := file.Read()
	if err != nil {
		return nil, err
	}

	instance := &RepositoryMemory{
		collection: impl.Merge(collection.DictionaryFromMap(requests)),
		file:       file,
	}

	go instance.watch()

	return instance, nil
}

func (r *RepositoryMemory) watch() {
	r.once.Do(func() {
		conf := configuration.Instance()
		if !conf.Snapshot().Enable {
			return
		}

		hub := make(chan system.SystemEvent, 1)
		defer close(hub)

		topics := []string{
			system.SNAPSHOT_TOPIC_END_POINT.TopicSnapshotApplyOutput(),
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

func (r *RepositoryMemory) read() error {
	requests, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.DictionaryFromMap(requests)
	return nil
}

func (r *RepositoryMemory) FindAll(owner string) []mock_domain.EndPointLite {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	filtered := r.collection.Filter(func(s string, e mock_domain.EndPoint) bool {
		return e.Owner == owner
	}).ValuesVector()

	return collection.VectorMap(filtered, func(e mock_domain.EndPoint) mock_domain.EndPointLite {
		return *mock_domain.LiteFromEndPoint(&e)
	}, collection.MakeVector).Collect()
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

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
