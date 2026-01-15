package collection

import (
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	collection_utils "github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection_utils.IDictionary[string, collection.Collection]
	file       repository.IFileManager[collection.Collection]
	close      chan bool
}

func InitializeRepositoryMemory(
	impl collection_utils.IDictionary[string, collection.Collection],
	file repository.IFileManager[collection.Collection]) (*RepositoryMemory, error) {
	collections, err := file.Read()
	if err != nil {
		return nil, err
	}
	instance := &RepositoryMemory{
		collection: impl.Merge(collection_utils.DictionaryFromMap(collections)),
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

		topics := []topic.TopicAction{
			topic_repository.TOPIC_COLLECTION.ActionReload(),
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

	r.collection = collection_utils.DictionaryFromMap(collections)
	return nil
}

func (r *RepositoryMemory) Find(id string) (*collection.Collection, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	collection, ok := r.collection.Get(id)
	return &collection, ok
}

func (r *RepositoryMemory) FindNodes(references []domain.NodeReference) []collection.NodeCollection {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	colls := make([]collection.NodeCollection, 0)
	for _, v := range references {
		coll, ok := r.collection.Get(v.Item)
		if !ok {
			continue
		}

		colls = append(colls, collection.NodeCollection{
			Order:      v.Order,
			Collection: coll,
		})
	}

	return colls
}

func (r *RepositoryMemory) Insert(owner string, collection *collection.Collection) *collection.Collection {
	r.muMemory.Lock()
	return r.resolve(owner, collection)
}

func (r *RepositoryMemory) resolve(owner string, collection *collection.Collection) *collection.Collection {
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

func (r *RepositoryMemory) insert(owner string, coll *collection.Collection) *collection.Collection {
	r.muMemory.Unlock()

	coll.Owner = owner

	if coll.Timestamp == 0 {
		coll.Timestamp = time.Now().UnixMilli()
	}

	coll.Modified = time.Now().UnixMilli()

	if coll.Name == "" {
		coll.Name = fmt.Sprintf("%s-%d", coll.Owner, coll.Timestamp)
	}

	if coll.Status == "" {
		coll.Status = collection.FREE
	}

	r.collection.Put(coll.Id, *coll)
	go r.write(r.collection)
	return coll
}

func (r *RepositoryMemory) Delete(collection *collection.Collection) *collection.Collection {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(collection.Id)
	go r.write(r.collection)

	return &cursor
}

func (r *RepositoryMemory) write(snapshot collection_utils.IDictionary[string, collection.Collection]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
