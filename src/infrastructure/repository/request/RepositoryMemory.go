package request

import (
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, action.Request]
	file       repository.IFileManager[action.Request]
	close      chan bool
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, action.Request], file repository.IFileManager[action.Request]) (*RepositoryMemory, error) {
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
			system.SNAPSHOT_TOPIC_REQUEST.TopicSnapshotApplyOutput(),
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

func (r *RepositoryMemory) Find(key string) (*action.Request, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(key)
}

func (r *RepositoryMemory) FindMany(ids ...string) []action.Request {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := make([]action.Request, 0)
	for _, v := range ids {
		if request, ok := r.collection.Get(v); ok {
			requests = append(requests, *request)
		}
	}

	return requests
}

func (r *RepositoryMemory) FindLiteNodes(references []domain.NodeReference) []dto.DtoLiteNodeRequest {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := make([]dto.DtoLiteNodeRequest, 0)
	for _, v := range references {
		if request, ok := r.collection.Get(v.Item); ok {
			requests = append(requests, dto.DtoLiteNodeRequest{
				Order:   v.Order,
				Request: *dto.ToLiteRequest(request),
			})
		}
	}

	return requests
}

func (r *RepositoryMemory) FindNodes(references []domain.NodeReference) []dto.DtoNodeRequest {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := make([]dto.DtoNodeRequest, 0)
	for _, v := range references {
		if request, ok := r.collection.Get(v.Item); ok {
			requests = append(requests, dto.DtoNodeRequest{
				Order:   v.Order,
				Request: *dto.FromRequest(request),
			})
		}
	}

	return requests
}

func (r *RepositoryMemory) FindRequests(references []domain.NodeReference) []action.Request {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()

	requests := make([]action.Request, 0)
	for _, v := range references {
		if request, ok := r.collection.Get(v.Item); ok {
			requests = append(requests, *request)
		}
	}

	return requests
}

func (r *RepositoryMemory) Insert(owner string, request *action.Request) *action.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	request.Owner = owner

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().UnixMilli()
	}

	request.Modified = time.Now().UnixMilli()

	if request.Name == "" {
		request.Name = fmt.Sprintf("%s-%s-%d", request.Owner, request.Method, request.Timestamp)
	}

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

func (r *RepositoryMemory) InsertMany(owner string, requests []action.Request) []action.Request {
	for i, v := range requests {
		req := r.Insert(owner, &v)
		requests[i] = *req
	}
	return requests
}

func (r *RepositoryMemory) Delete(request *action.Request) *action.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(request.Id)
	if cursor != nil {
		go r.write(r.collection)
	}

	return cursor
}

func (r *RepositoryMemory) DeleteMany(requests ...action.Request) []action.Request {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	deleted := make([]action.Request, 0)
	for _, v := range requests {
		cursor, _ := r.collection.Remove(v.Id)
		deleted = append(deleted, *cursor)
	}

	go r.write(r.collection)

	return deleted
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, action.Request]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
