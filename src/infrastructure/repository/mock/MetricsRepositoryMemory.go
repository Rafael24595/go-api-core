package mock

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

const NameMetricsMemory = "metrics_memory" 

type MetricsRepositoryMemory struct {
	once       sync.Once
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, mock_domain.Metrics]
	file       repository.IFileManager[mock_domain.Metrics]
	close      chan bool
}

func InitializeMetricsRepositoryMemory(
	impl collection.IDictionary[string, mock_domain.Metrics],
	file repository.IFileManager[mock_domain.Metrics]) (*MetricsRepositoryMemory, error) {
	raw, err := file.Read()
	if err != nil {
		return nil, err
	}

	metrics := impl.Merge(collection.DictionaryFromMap(raw))

	instance := &MetricsRepositoryMemory{
		collection: metrics,
		file:       file,
	}

	go instance.watch()

	return instance, nil
}

func (r *MetricsRepositoryMemory) watch() {
	r.once.Do(func() {
		conf := configuration.Instance()
		if !conf.Snapshot().Enable {
			return
		}

		hub := make(chan system.SystemEvent, 1)
		defer close(hub)

		topics := []topic.TopicAction{
			topic_repository.TOPIC_METRICS.ActionReload(),
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
				log.Customf(repository.RepositoryCategory, "The repository %q has been reloaded.", NameMetricsMemory)
			case <-conf.Signal.Done():
				log.Customf(repository.RepositoryCategory, "Watcher stopped: global shutdown signal received.")
				return
			}
		}
	})
}

func (r *MetricsRepositoryMemory) read() error {
	requests, err := r.file.Read()
	if err != nil {
		return err
	}

	r.collection = collection.DictionaryFromMap(requests)
	return nil
}

func (r *MetricsRepositoryMemory) Find(endPoint *mock_domain.EndPoint) (*mock_domain.Metrics, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	metrics, ok := r.collection.Get(endPoint.Id)
	return &metrics, ok
}

func (r *MetricsRepositoryMemory) Resolve(endPoint *mock_domain.EndPoint, metrics *mock_domain.Metrics) *mock_domain.Metrics {
	if endPoint.Id == "" || metrics.EndPoint != "" && metrics.EndPoint != endPoint.Id {
		return nil
	}

	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	metrics.EndPoint = endPoint.Id

	now := time.Now().UnixMilli()
	if metrics.Timestamp == 0 {
		metrics.Timestamp = now
	}

	metrics.Modified = now

	r.collection.Put(metrics.EndPoint, *metrics)

	go r.write(r.collection)

	return metrics
}

func (r *MetricsRepositoryMemory) Delete(endPoint *mock_domain.EndPoint, metrics *mock_domain.Metrics) *mock_domain.Metrics {
	if endPoint.Id == "" || metrics.EndPoint != "" && metrics.EndPoint != endPoint.Id {
		return nil
	}

	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(endPoint.Id)
	go r.write(r.collection)

	return &cursor
}

func (r *MetricsRepositoryMemory) write(snapshot collection.IDictionary[string, mock_domain.Metrics]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	err := r.file.Write(snapshot.Values())
	if err != nil {
		log.Error(err)
	}
}
