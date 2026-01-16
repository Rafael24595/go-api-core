package repository

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-api-core/src/domain/session"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

const NameMemory = "session_memory"

var (
	once sync.Once
)

type RepositoryMemory struct {
	once     sync.Once
	mut      sync.RWMutex
	mutFile  sync.RWMutex
	file     repository.IFileManager[dto.DtoSession]
	sessions collection.IDictionary[string, session.Session]
	close    chan bool
}

func InitializeRepositoryMemory(file repository.IFileManager[dto.DtoSession]) (*RepositoryMemory, error) {
	var instance *RepositoryMemory
	var err error

	once.Do(func() {
		steps, res := file.Read()
		if res != nil {
			err = res
			return
		}

		sessions := collection.MapToDictionarySync(steps,
			func(k string, d dto.DtoSession) session.Session {
				return *dto.ToSession(d)
			})

		instance = &RepositoryMemory{
			sessions: sessions,
			file:     file,
		}

		go instance.watch()
	})

	return instance, err
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
			topic_repository.TOPIC_SESSION.ActionReload(),
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
					log.Custome(repository.SnapshotCategory, err)
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
	sessions, err := r.file.Read()
	if err != nil {
		return err
	}

	r.sessions = collection.DictionarySyncMap(
		collection.DictionarySyncFromMap(sessions),
		func(k string, d dto.DtoSession) session.Session {
			return *dto.ToSession(d)
		})

	return nil
}

func (r *RepositoryMemory) FindAll() []session.Session {
	return r.sessions.Values()
}

func (r *RepositoryMemory) Find(user string) (*session.Session, bool) {
	session, ok := r.sessions.Get(user)
	return &session, ok
}

func (r *RepositoryMemory) Insert(sess *session.Session) *session.Session {
	r.mut.Lock()
	defer r.mut.Unlock()

	r.sessions.Put(sess.Username, *sess)

	go r.write(r.sessions)

	return sess
}

func (r *RepositoryMemory) Delete(sess *session.Session) *session.Session {
	r.mut.RLock()
	defer r.mut.RUnlock()

	deleted, ok := r.sessions.Remove(sess.Username)
	if ok {
		go r.write(r.sessions)
	}

	return &deleted
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, session.Session]) {
	r.mutFile.Lock()
	defer r.mutFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v session.Session) dto.DtoSession {
		return *dto.FromSession(v)
	})

	err := r.file.Write(items.Values())
	if err != nil {
		log.Error(err)
	}
}
