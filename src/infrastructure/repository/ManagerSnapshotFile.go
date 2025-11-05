package repository

import (
	"fmt"
	"maps"
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
const snapshotTopic = "snapshot"

const snpsh = "snpsh"
const snpsh_timestamp = `^(snpsh_)(\d*)(\.csvt)$`

const snapshotRetries = 3
const snapshotRetrySeconds = 30

type ManagerSnapshotFile[T IStructure] struct {
	once    sync.Once
	close   chan bool
	limit   int
	path    string
	errors  []error
	time    int64
	last    int64
	manager IFileManager[T]
}

func InitializeManagerSnapshotFile[T IStructure](path string, time int64, limit int, manager IFileManager[T]) *ManagerSnapshotFile[T] {
	if limit < 1 {
		limit = 1
	}

	instance := &ManagerSnapshotFile[T]{
		limit:   limit,
		path:    path,
		errors:  make([]error, 0),
		time:    time,
		last:    0,
		manager: manager,
	}

	go instance.watch(time)

	return instance
}

func (m *ManagerSnapshotFile[T]) watch(millis int64) {
	m.once.Do(func() {
		conf := configuration.Instance()
		ticker := time.NewTicker(time.Duration(millis) * time.Millisecond)
		defer ticker.Stop()

		hub := make(chan system.SystemEvent, 1)
		defer close(hub)

		conf.EventHub.Subcribe(snapshotListener, hub, snapshotTopic)
		defer conf.EventHub.Unsubcribe(snapshotListener)

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

func (m *ManagerSnapshotFile[T]) unwatch() {
	m.close <- true
}

func (m *ManagerSnapshotFile[T]) trySnapshot() {
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

func (m *ManagerSnapshotFile[T]) snapshot() error {
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

func (m *ManagerSnapshotFile[T]) collect() (*collection.Vector[os.DirEntry], error) {
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

func (m *ManagerSnapshotFile[T]) save(name string) error {
	snapshot, err := m.manager.Read()
	if err != nil {
		return err
	}

	items := make([]any, 0)
	values := maps.Values(snapshot)
	for v := range values {
		items = append(items, v)
	}

	result, err := m.manager.marshal(items)
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

func (m *ManagerSnapshotFile[T]) clean(snapshots collection.Vector[os.DirEntry]) error {
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

func (m *ManagerSnapshotFile[T]) Read() (map[string]T, error) {
	return m.manager.Read()
}

func (m *ManagerSnapshotFile[T]) Write(items []any) error {
	return m.manager.Write(items)
}

func (m *ManagerSnapshotFile[T]) marshal(items []any) ([]byte, error) {
	return m.manager.marshal(items)
}
