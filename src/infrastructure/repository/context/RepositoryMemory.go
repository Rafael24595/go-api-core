package historic

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, context.Context]
	file       repository.IFileManager[dto.DtoContext]
	close      chan bool
}

func InitializeRepositoryMemory(
	impl collection.IDictionary[string, context.Context],
	file repository.IFileManager[dto.DtoContext]) (*RepositoryMemory, error) {
	steps, err := file.Read()
	if err != nil {
		return nil, err
	}

	ctx := collection.MapToDictionary(steps,
		func(k string, d dto.DtoContext) context.Context {
			return *dto.ToContext(&d)
		})

	instance := &RepositoryMemory{
		collection: impl.Merge(ctx),
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
			system.SNAPSHOT_TOPIC_CONTEXT.TopicSnapshotApplyOutput(),
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
	ctx, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.MapToDictionary(ctx,
		func(k string, d dto.DtoContext) context.Context {
			return *dto.ToContext(&d)
		})

	return nil
}

func (r *RepositoryMemory) Find(id string) (*context.Context, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	context, ok := r.collection.Get(id)
	return &context, ok
}

func (r *RepositoryMemory) Insert(owner, collection string, ctx *context.Context) *context.Context {
	return r.resolve(owner, collection, ctx)
}

func (r *RepositoryMemory) resolve(owner, collection string, ctx *context.Context) *context.Context {
	if ctx.Id != "" {
		return r.insert(owner, collection, ctx)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, collection, ctx)
	}

	ctx.Id = key

	return r.insert(owner, collection, ctx)
}

func (r *RepositoryMemory) insert(owner, collection string, ctx *context.Context) *context.Context {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	ctx.Owner = owner
	ctx.Collection = collection

	if ctx.Timestamp == 0 {
		ctx.Timestamp = time.Now().UnixMilli()
	}

	ctx.Modified = time.Now().UnixMilli()

	r.collection.Put(ctx.Id, *ctx)

	go r.write(r.collection)

	return ctx
}

func (r *RepositoryMemory) Update(owner string, ctx *context.Context) (*context.Context, bool) {
	if _, exists := r.Find(ctx.Id); !exists {
		return ctx, false
	}
	return r.resolve(owner, ctx.Collection, ctx), true
}

func (r *RepositoryMemory) Delete(context *context.Context) *context.Context {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(context.Id)
	go r.write(r.collection)

	return &cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, context.Context]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v context.Context) dto.DtoContext {
		return *dto.FromContext(&v)
	})

	err := r.file.Write(items.Values())
	if err != nil {
		log.Error(err)
	}
}
