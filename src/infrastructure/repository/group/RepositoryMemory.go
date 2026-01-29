package historic

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-api-core/src/domain/group"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

const NameMemory = "group_memory"

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, group.Group]
	file       repository.IFileManager[group.Group]
	close      chan bool
}

func InitializeRepositoryMemory(
	impl collection.IDictionary[string, group.Group],
	file repository.IFileManager[group.Group]) (*RepositoryMemory, error) {
	groups, err := file.Read()
	if err != nil {
		return nil, err
	}

	instance := &RepositoryMemory{
		collection: impl.Merge(collection.DictionaryFromMap(groups)),
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
			topic_repository.TOPIC_GROUP.ActionReload(),
		}

		conf.EventHub.Subcribe(repository.RepositoryListener, hub, topics...)
		defer conf.EventHub.Unsubcribe(repository.RepositoryListener, topics...)

		for {
			select {
			case <-r.close:
				log.Customf(repository.RepositoryCategory, "Watcher stopped: local close signal received.")
				return
			case <-hub:
				if err := r.read(); err != nil {
					log.Custome(repository.RepositoryCategory, err)
					return
				}
				log.Customf(repository.RepositoryCategory, "The repository %q has been reloaded.", NameMemory)
			case <-conf.Signal.Done():
				log.Customf(repository.RepositoryCategory, "Watcher stopped: global shutdown signal received.")
				return
			}
		}
	})
}

func (r *RepositoryMemory) read() error {
	groups, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.DictionaryFromMap(groups)
	return nil
}

func (r *RepositoryMemory) Find(id string) (*group.Group, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	group, ok := r.collection.Get(id)
	return &group, ok
}

func (r *RepositoryMemory) Insert(owner string, group *group.Group) *group.Group {
	r.muMemory.Lock()
	return r.resolve(owner, group)
}

func (r *RepositoryMemory) resolve(owner string, group *group.Group) *group.Group {
	if group.Id != "" {
		return r.insert(owner, group)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, group)
	}

	group.Id = key

	return r.insert(owner, group)
}

func (r *RepositoryMemory) insert(owner string, group *group.Group) *group.Group {
	defer r.muMemory.Unlock()

	group.Owner = owner

	if group.Timestamp == 0 {
		group.Timestamp = time.Now().UnixMilli()
	}

	group.Modified = time.Now().UnixMilli()

	r.collection.Put(group.Id, *group)

	go r.write(r.collection)

	return group
}

func (r *RepositoryMemory) Delete(context *group.Group) *group.Group {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(context.Id)
	go r.write(r.collection)

	return &cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, group.Group]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
