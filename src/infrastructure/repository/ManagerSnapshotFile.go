package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	command_helper "github.com/Rafael24595/go-api-core/src/commons/command/helper"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const snapshotCategory = "SNAPSHOT"

const SnapshotListener = "snapshot"

const snpsh = "snpsh"

const snapshotRetries = 3
const snapshotRetrySeconds = 30

type builderManagerSnapshotFile[T IStructure] struct {
	manager *managerSnapshotFile[T]
}

func BuilderManagerSnapshotFile[T IStructure](topic system.TopicSnapshot, manager IFileManager[T]) *builderManagerSnapshotFile[T] {
	instance := &managerSnapshotFile[T]{
		limit:   1,
		topic:   topic,
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
	topic   system.TopicSnapshot
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

		topics := []string{
			m.topic.TopicSnapshotSaveInput(),
			m.topic.TopicSnapshotAppyInput(),
		}

		conf.EventHub.Subcribe(SnapshotListener, hub, topics...)
		defer conf.EventHub.Unsubcribe(SnapshotListener, topics...)

		m.trySnapshot(false, conf.Format())

		for {
			select {
			case <-m.close:
				log.Customf(snapshotCategory, "Watcher stopped: local close signal received.")
				return
			case e := <-hub:
				switch e.Topic {
				case m.topic.TopicSnapshotSaveInput():
					m.trySave(e)
				case m.topic.TopicSnapshotAppyInput():
					m.tryApply(e)
				}
			case <-conf.Signal.Done():
				log.Customf(snapshotCategory, "Watcher stopped: global shutdown signal received.")
				return
			case <-ticker.C:
				m.trySnapshot(false, conf.Format())
			}
		}
	})
}

func (m *managerSnapshotFile[T]) trySave(e system.SystemEvent) error {
	format, ok := format.DataFormatFromString(e.Value.String())
	if !ok {
		return fmt.Errorf("unsupported format extension %q", e.Value.String())
	}

	return m.snapshot(true, format)
}

func (m *managerSnapshotFile[T]) tryApply(e system.SystemEvent) error {
	file := e.Value.String()

	ext := filepath.Ext(file)
	if ext == "" {
		return fmt.Errorf("undefined extension for snapshot %q", file)
	}

	format, ok := format.DataFormatFromExtension(ext)
	if !ok {
		return fmt.Errorf("unsupported format extension %q", ext)
	}

	snapshots, err := m.collect(format)
	if err != nil {
		return err
	}

	cursor, ok := snapshots.FindOne(func(f os.DirEntry) bool {
		return strings.HasPrefix(f.Name(), file)
	})

	if !ok {
		return nil
	}

	return m.apply((*cursor).Name(), format)
}

func (m *managerSnapshotFile[T]) unwatch() {
	m.close <- true
}

func (m *managerSnapshotFile[T]) trySnapshot(force bool, format format.DataFormat) {
	ticker := time.NewTicker(snapshotRetrySeconds * time.Second)
	defer ticker.Stop()

	for {
		err := m.snapshot(force, format)
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

func (m *managerSnapshotFile[T]) snapshot(force bool, format format.DataFormat) error {
	snapshots, err := m.collect(format)
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()

	if m.last == 0 {
		last, ok := snapshots.Last()
		if ok && last != nil {
			re := regexp.MustCompile(command_helper.SnpshTimestamp)
			rawLast := re.FindStringSubmatch((*last).Name())[2]
			timeLast, _ := strconv.ParseInt(rawLast, 10, 64)
			m.last = timeLast
		} else {
			m.last = now - m.time
		}
	}

	if force || now-m.last >= m.time {
		extension := format.Extension()
		code := fmt.Sprintf("%s_%d.%s", snpsh, now, extension)
		m.save(code, format)
	}

	snapshots, err = m.collect(format)
	if err != nil {
		return err
	}

	err = m.clean(*snapshots, format)
	if err != nil {
		return err
	}

	return nil
}

func (m *managerSnapshotFile[T]) collect(format format.DataFormat) (*collection.Vector[os.DirEntry], error) {
	path, err := m.path(format)
	if err != nil {
		return nil, err
	}

	return command_helper.FindSnapshots(path)
}

func (m *managerSnapshotFile[T]) save(name string, format format.DataFormat) error {
	snapshot, err := m.manager.Read()
	if err != nil {
		return err
	}

	items := collection.DictionaryFromMap(snapshot)

	result, err := TryMarshal(format, items.Values())
	if err != nil {
		return err
	}

	path, err := m.path(format)
	if err != nil {
		return err
	}

	location := filepath.Join(path, name)
	err = utils.WriteFile(location, string(result))
	if err == nil {
		log.Customf(snapshotCategory, "A new snapshot %q has been defined.", location)
	}

	return err
}

func (m *managerSnapshotFile[T]) apply(name string, format format.DataFormat) error {
	path, err := m.path(format)
	if err != nil {
		return err
	}
	
	location := filepath.Join(path, name)
	buffer, err := utils.ReadFile(location)
	if err != nil {
		return err
	}

	snapshot, err := TryUnmarshal[T](format, buffer)
	if err != nil {
		return err
	}

	items := collection.DictionaryFromMap(snapshot)

	err = m.manager.Write(items.Values())
	if err != nil {
		return err
	}

	conf := configuration.Instance()
	conf.EventHub.Publish(m.topic.TopicSnapshotApplyOutput(), "true")

	return nil
}

func (m *managerSnapshotFile[T]) clean(snapshots collection.Vector[os.DirEntry], format format.DataFormat) error {
	size := snapshots.Size()
	if size == 0 || size < m.limit {
		return nil
	}

	path, err := m.path(format)
	if err != nil {
		return err
	}

	for snapshots.Size() > m.limit {
		cursor, ok := snapshots.Shift()
		if !ok {
			return nil
		}

		path := filepath.Join(path, (*cursor).Name())
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}

		log.Customf(snapshotCategory, "An old snapshot %q has been removed.", path)
	}

	return nil
}

func (m *managerSnapshotFile[T]) path(format format.DataFormat) (string, error) {
	path, ok := m.topic.Path(format)
	if !ok {
		return "", fmt.Errorf("unsupported format %q", format)
	}
	return path, nil
}

func (m *managerSnapshotFile[T]) Read() (map[string]T, error) {
	return m.manager.Read()
}

func (m *managerSnapshotFile[T]) Write(items []T) error {
	return m.manager.Write(items)
}
