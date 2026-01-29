package token

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	token_domain "github.com/Rafael24595/go-api-core/src/domain/token"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

const NameMemory = "token_memory" 

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, token_domain.Token]
	file       repository.IFileManager[token_domain.Token]
	close      chan bool
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, token_domain.Token], file repository.IFileManager[token_domain.Token]) (*RepositoryMemory, error) {
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

		topics := []topic.TopicAction{
			topic_repository.TOPIC_TOKEN.ActionReload(),
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
	requests, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.DictionaryFromMap(requests)
	return nil
}

func (r *RepositoryMemory) FindAll(owner string) []token_domain.LiteToken {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	tokens := r.collection.ValuesVector().
		Filter(func(t token_domain.Token) bool {
			return t.Owner == owner
		})
	return collection.VectorMap(tokens, func(t token_domain.Token) token_domain.LiteToken {
		return token_domain.ToLiteToken(t)
	}).Collect()
}

func (r *RepositoryMemory) Find(id string) (*token_domain.Token, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	token, ok := r.collection.Get(id)
	return &token, ok
}

func (r *RepositoryMemory) FindByName(owner, name string) (*token_domain.Token, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	token, ok := r.collection.FindOne(func(s string, t token_domain.Token) bool {
		return t.Owner == owner && t.Name == name
	})
	return &token, ok
}

func (r *RepositoryMemory) FindByToken(owner, token string) (*token_domain.Token, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	tkn, ok := r.collection.FindOne(func(s string, t token_domain.Token) bool {
		return t.Owner == owner && t.Token == token
	})
	return &tkn, ok
}

func (r *RepositoryMemory) FindGlobal(token string) (*token_domain.Token, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	tkn, ok := r.collection.FindOne(func(s string, t token_domain.Token) bool {
		return t.Token == token
	})
	return &tkn, ok
}

func (r *RepositoryMemory) Insert(owner string, token *token_domain.Token) *token_domain.Token {
	r.muMemory.Lock()
	return r.resolve(owner, token)
}

func (r *RepositoryMemory) resolve(owner string, token *token_domain.Token) *token_domain.Token {
	if token.Id != "" {
		return r.insert(owner, token)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, token)
	}

	token.Id = key

	return r.insert(owner, token)
}

func (r *RepositoryMemory) insert(owner string, token *token_domain.Token) *token_domain.Token {
	defer r.muMemory.Unlock()

	token.Owner = owner

	if token.Timestamp == 0 {
		token.Timestamp = time.Now().UnixMilli()
	}

	r.collection.Put(token.Id, *token)

	go r.write(r.collection)

	return token
}

func (r *RepositoryMemory) Delete(endPoint *token_domain.Token) *token_domain.Token {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(endPoint.Id)
	go r.write(r.collection)

	return &cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, token_domain.Token]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
