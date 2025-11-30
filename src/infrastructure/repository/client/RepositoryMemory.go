package historic

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/domain/client"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, client.ClientData]
	file       repository.IFileManager[client.ClientData]
	close      chan bool
}

func InitializeRepositoryMemory(
	impl collection.IDictionary[string, client.ClientData],
	file repository.IFileManager[client.ClientData]) (*RepositoryMemory, error) {
	raw, err := file.Read()
	if err != nil {
		return nil, err
	}

	data := collection.DictionaryFromMap(raw)

	instance := &RepositoryMemory{
		collection: impl.Merge(data),
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
			system.SNAPSHOT_TOPIC_CLIENT_DATA.TopicSnapshotApplyOutput(),
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
	raw, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.DictionaryFromMap(raw)

	return nil
}

func (r *RepositoryMemory) Find(owner string) (*client.ClientData, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(owner)
}

func (r *RepositoryMemory) Insert(data *client.ClientData) *client.ClientData {
	return r.insert(data)
}

func (r *RepositoryMemory) insert(data *client.ClientData) *client.ClientData {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	if data.Owner == "" {
		return nil
	}

	if data.Timestamp == 0 {
		data.Timestamp = time.Now().UnixMilli()
	}

	data.Modified = time.Now().UnixMilli()

	r.collection.Put(data.Owner, *data)

	go r.write(r.collection)

	return data
}

func (r *RepositoryMemory) Update(data *client.ClientData) (*client.ClientData, bool) {
	if _, exists := r.Find(data.Owner); !exists {
		return data, false
	}
	return r.insert(data), true
}

func (r *RepositoryMemory) Delete(data *client.ClientData) *client.ClientData {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(data.Owner)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, client.ClientData]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
