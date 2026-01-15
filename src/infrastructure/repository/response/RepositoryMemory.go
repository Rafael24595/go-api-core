package response

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, action.Response]
	file       repository.IFileManager[action.Response]
	close      chan bool
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, action.Response], file repository.IFileManager[action.Response]) (*RepositoryMemory, error) {
	responses, err := file.Read()
	if err != nil {
		return nil, err
	}

	instance := &RepositoryMemory{
		collection: impl.Merge(collection.DictionaryFromMap(responses)),
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
			topic_repository.TOPIC_RESPONSE.ActionReload(),
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

func (r *RepositoryMemory) Find(key string) (*action.Response, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	response, ok := r.collection.Get(key)
	return &response, ok
}

func (r *RepositoryMemory) FindMany(ids []string) []action.Response {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	responses := make([]action.Response, 0)
	for _, v := range ids {
		if response, ok := r.collection.Get(v); ok {
			responses = append(responses, response)
		}
	}

	return responses
}

func (r *RepositoryMemory) Insert(owner string, response *action.Response) *action.Response {
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

func (r *RepositoryMemory) Delete(response *action.Response) *action.Response {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(response.Id)
	go r.write(r.collection)

	return &cursor
}

func (r *RepositoryMemory) DeleteMany(responses ...action.Response) []action.Response {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	deleted := make([]action.Response, 0)
	for _, v := range responses {
		cursor, _ := r.collection.Remove(v.Id)
		deleted = append(deleted, cursor)
	}

	go r.write(r.collection)

	return deleted
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, action.Response]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
