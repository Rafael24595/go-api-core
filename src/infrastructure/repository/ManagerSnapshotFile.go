package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const snapshotCategory = "SNAPSHOT"

const snapshotListener = "snapshot"

const snpsh = "snpsh"
const snpsh_timestamp = `^(snpsh_)(\d*)(\.csvt)$`

const snapshotRetries = 3
const snapshotRetrySeconds = 30

type builderManagerSnapshotFile[T IStructure] struct {
	manager *managerSnapshotFile[T]
}

func BuilderManagerSnapshotFile[T IStructure](path, topic string, manager IFileManager[T]) *builderManagerSnapshotFile[T] {
	instance := &managerSnapshotFile[T]{
		limit:   1,
		topic:   topic,
		path:    path,
		errors:  make([]error, 0),
		time:    int64(time.Hour) * 24 * 7,
		last:    0,
		manager: manager,
	}

	return &builderManagerSnapshotFile[T]{
		manager: instance,
	}
}

func (b *builderManagerSnapshotFile[T]) Limit(limit int) *builderManagerSnapshotFile[T] {
	if limit < 1 {
		limit = 1
	}

	b.manager.limit = limit
	return b
}

func (b *builderManagerSnapshotFile[T]) Time(time int64) *builderManagerSnapshotFile[T] {
	if time < 0 {
		return b
	}

	b.manager.time = time
	return b
}

func (b *builderManagerSnapshotFile[T]) Make() *managerSnapshotFile[T] {
	b.manager.Initialize()
	return b.manager
}

type managerSnapshotFile[T IStructure] struct {
	once    sync.Once
	close   chan bool
	limit   int
	path    string
	topic   string
	errors  []error
	time    int64
	last    int64
	manager IFileManager[T]
}

func (m *managerSnapshotFile[T]) Initialize() {
	go m.watch()
}

func (m *managerSnapshotFile[T]) watch() {
	m.once.Do(func() {
		conf := configuration.Instance()
		ticker := time.NewTicker(time.Duration(m.time))
		defer ticker.Stop()

		hub := make(chan system.SystemEvent, 1)
		defer close(hub)

		conf.EventHub.Subcribe(snapshotListener, hub, m.topic)
		defer conf.EventHub.Unsubcribe(snapshotListener, m.topic)

		m.trySnapshot()

		for {
			select {
			case <-m.close:
				log.Customf(snapshotCategory, "Watcher stopped: local close signal received.")
				return
			case e := <-hub:
				//TODO: Implement.
				fmt.Printf("TODO: Use: %q", e.Value.String())
			case <-conf.Signal.Done():
				log.Customf(snapshotCategory, "Watcher stopped: global shutdown signal received.")
				return
			case <-ticker.C:
				m.trySnapshot()
			}
		}
	})
}

func (m *managerSnapshotFile[T]) unwatch() {
	m.close <- true
}

func (m *managerSnapshotFile[T]) trySnapshot() {
	ticker := time.NewTicker(snapshotRetrySeconds * time.Second)
	defer ticker.Stop()

	for {
		err := m.snapshot()
		if err != nil {
			log.Error(err)
			m.errors = append(m.errors, err)

			if len(m.errors) >= snapshotRetries {
				log.Errors("The maximum error rate has been exceeded; snapshot generation will be discontinued.")
				m.unwatch()
				return
			}

			<-ticker.C
			continue
		}

		m.errors = make([]error, 0)

		return
	}
}

func (m *managerSnapshotFile[T]) snapshot() error {
	snapshots, err := m.collect()
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()

	if m.last == 0 {
		last, ok := snapshots.Last()
		if ok && last != nil {
			re := regexp.MustCompile(snpsh_timestamp)
			rawLast := re.FindStringSubmatch((*last).Name())[2]
			timeLast, _ := strconv.ParseInt(rawLast, 10, 64)
			m.last = timeLast
		} else {
			m.last = now - m.time
		}
	}

	if now-m.last >= m.time {
		name := fmt.Sprintf("%s_%d", snpsh, now)
		m.save(name)
	}

	snapshots, err = m.collect()
	if err != nil {
		return err
	}

	err = m.clean(*snapshots)
	if err != nil {
		return err
	}

	return nil
}

func (m *managerSnapshotFile[T]) collect() (*collection.Vector[os.DirEntry], error) {
	err := os.MkdirAll(m.path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadDir(m.path)
	if err != nil {
		return nil, fmt.Errorf("an error ocurred during snapshot directory %q reading: %s", m.path, err.Error())
	}

	re := regexp.MustCompile(snpsh_timestamp)

	return collection.VectorFromList(raw).
		Filter(func(d os.DirEntry) bool {
			return !d.IsDir() && len(re.FindStringSubmatch(d.Name())) == 4
		}).
		Sort(func(a, b os.DirEntry) bool {
			ar := re.FindStringSubmatch(a.Name())[2]
			at, _ := strconv.ParseInt(ar, 10, 64)

			br := re.FindStringSubmatch(b.Name())[2]
			bt, _ := strconv.ParseInt(br, 10, 64)

			return at < bt
		}), nil
}

func (m *managerSnapshotFile[T]) save(name string) error {
	snapshot, err := m.manager.Read()
	if err != nil {
		return err
	}

	items := collection.DictionaryFromMap(snapshot)

	result, err := m.manager.marshal(items.Values())
	if err != nil {
		return err
	}

	path := filepath.Join(m.path, fmt.Sprintf("%s.csvt", name))
	err = utils.WriteFile(path, string(result))
	if err != nil {
		log.Customf(snapshotCategory, "A new snapshot %q has been defined.", path)
	}

	return err
}

func (m *managerSnapshotFile[T]) apply(name string) error {
	path := filepath.Join(m.path, fmt.Sprintf("%s.csvt", name))
	buffer, err := utils.ReadFile(path)
	if err != nil {
		return err
	}

	snapshot, err := m.manager.unmarshal(buffer)
	if err != nil {
		return err
	}

	items := collection.DictionaryFromMap(snapshot)

	err = m.manager.Write(items.Values())
	if err != nil {
		return err
	}

	conf := configuration.Instance()
	conf.EventHub.Publish("//TODO: Refresh repository.", "//TODO: Evalue any as payload.")

	return nil
}

func (m *managerSnapshotFile[T]) clean(snapshots collection.Vector[os.DirEntry]) error {
	size := snapshots.Size()
	if size == 0 || size < m.limit {
		return nil
	}

	for snapshots.Size() > m.limit {
		cursor, ok := snapshots.Shift()
		if !ok {
			return nil
		}

		path := filepath.Join(m.path, (*cursor).Name())
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}

		log.Customf(snapshotCategory, "An old snapshot %q has been removed.", path)
	}

	return nil
}

func (m *managerSnapshotFile[T]) Read() (map[string]T, error) {
	return m.manager.Read()
}

func (m *managerSnapshotFile[T]) Write(items []T) error {
	return m.manager.Write(items)
}

func (m *managerSnapshotFile[T]) unmarshal(buffer []byte) (map[string]T, error) {
	return m.manager.unmarshal(buffer)
}

func (m *managerSnapshotFile[T]) marshal(items []T) ([]byte, error) {
	return m.manager.marshal(items)
}
