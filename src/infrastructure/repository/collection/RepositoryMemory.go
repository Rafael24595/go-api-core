package collection

import (
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	collection_domain "github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, collection_domain.Collection]
	file       repository.IFileManager[collection_domain.Collection]
	close      chan bool
}

func InitializeRepositoryMemory(
	impl collection.IDictionary[string, collection_domain.Collection],
	file repository.IFileManager[collection_domain.Collection]) (*RepositoryMemory, error) {
	collections, err := file.Read()
	if err != nil {
		return nil, err
	}
	instance := &RepositoryMemory{
		collection: impl.Merge(collection.DictionaryFromMap(collections)),
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
			system.SNAPSHOT_TOPIC_COLLECTION.TopicSnapshotApplyOutput(),
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
	collections, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.DictionaryFromMap(collections)
	return nil
}

func (r *RepositoryMemory) Find(id string) (*collection_domain.Collection, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(id)
}

func (r *RepositoryMemory) FindOneBystatus(owner string, status collection_domain.StatusCollection) (*collection_domain.Collection, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.ValuesVector().
		FindOne(func(c collection_domain.Collection) bool {
			return c.Owner == owner && c.Status == status
		})
}

func (r *RepositoryMemory) FindAllBystatus(owner string, status collection_domain.StatusCollection) []collection_domain.Collection {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.ValuesVector().
		Filter(func(c collection_domain.Collection) bool {
			return c.Owner == owner && c.Status == status
		}).
		Collect()
}

func (r *RepositoryMemory) FindCollections(references []domain.NodeReference) []collection_domain.NodeCollection {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	collections := make([]collection_domain.NodeCollection, 0)
	for _, v := range references {
		if collection, ok := r.collection.Get(v.Item); ok {
			collections = append(collections, collection_domain.NodeCollection{
				Order:      v.Order,
				Collection: *collection,
			})
		}
	}

	return collections
}

func (r *RepositoryMemory) Insert(owner string, collection *collection_domain.Collection) *collection_domain.Collection {
	r.muMemory.Lock()
	return r.resolve(owner, collection)
}

func (r *RepositoryMemory) resolve(owner string, collection *collection_domain.Collection) *collection_domain.Collection {
	if collection.Id != "" {
		return r.insert(owner, collection)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, collection)
	}

	collection.Id = key

	return r.insert(owner, collection)
}

func (r *RepositoryMemory) insert(owner string, collection *collection_domain.Collection) *collection_domain.Collection {
	r.muMemory.Unlock()

	collection.Owner = owner

	if collection.Timestamp == 0 {
		collection.Timestamp = time.Now().UnixMilli()
	}

	collection.Modified = time.Now().UnixMilli()

	if collection.Name == "" {
		collection.Name = fmt.Sprintf("%s-%d", collection.Owner, collection.Timestamp)
	}

	if collection.Status == "" {
		collection.Status = collection_domain.FREE
	}

	r.collection.Put(collection.Id, *collection)
	go r.write(r.collection)
	return collection
}

func (r *RepositoryMemory) PushToCollection(owner string, collection *collection_domain.Collection, request *action.Request) (*collection_domain.Collection, *action.Request) {
	r.muMemory.Lock()

	if !collection.ExistsRequest(request.Id) {
		collection.Nodes = append(collection.Nodes, domain.NodeReference{
			Order: len(collection.Nodes),
			Item:  request.Id,
		})
	}

	return r.resolve(owner, collection), request
}

func (r *RepositoryMemory) Delete(collection *collection_domain.Collection) *collection_domain.Collection {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(collection.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, collection_domain.Collection]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
