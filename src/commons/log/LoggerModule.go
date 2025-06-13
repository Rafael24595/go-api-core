package log

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Module interface {
	Name() string
	Vector([]Record) []Record
	Record(*Record, bool) *Record
}

const CODE_LOGGER_MODULE = "MODULE"

type loggerModule struct {
	mu            sync.Mutex
	loggerModules []Module
	records       []Record
}

func newloggerModule() *loggerModule {
	return &loggerModule{
		loggerModules: make([]Module, 0),
		records:       make([]Record, 0),
	}
}

func (l *loggerModule) pushModule(module Module) *loggerModule {
	l.loggerModules = append(l.loggerModules, module)
	return l
}

func (l *loggerModule) Name() string {
	return CODE_LOGGER_MODULE
}

func (l *loggerModule) Metadata() string {
	modules := make([]string, len(l.loggerModules))
	for i, v := range l.loggerModules {
		modules[i] = v.Name()
	}
	return fmt.Sprintf("Modules[%d]: %s", len(l.loggerModules), strings.Join(modules, ", "))
}

func (l *loggerModule) Records() []Record {
	l.mu.Lock()
	defer l.mu.Unlock()

	clone := make([]Record, len(l.records))
	copy(clone, l.records)

	return clone
}

func (l *loggerModule) Message(message string) *Record {
	return l.record(MESSAGE, message, false)
}

func (l *loggerModule) Messagef(format string, args ...any) *Record {
	return l.record(MESSAGE, fmt.Sprintf(format, args...), false)
}

func (l *loggerModule) Warning(message string) *Record {
	return l.record(WARNING, message, false)
}

func (l *loggerModule) Warningf(format string, args ...any) *Record {
	return l.record(WARNING, fmt.Sprintf(format, args...), false)
}

func (l *loggerModule) Error(err error) *Record {
	return l.record(ERROR, err.Error(), false)
}

func (l *loggerModule) Errors(message string) *Record {
	return l.record(ERROR, message, false)
}

func (l *loggerModule) Errorf(format string, args ...any) *Record {
	return l.record(ERROR, fmt.Sprintf(format, args...), false)
}

func (l *loggerModule) Panic(err error) {
	l.record(PANIC, err.Error(), true)
}

func (l *loggerModule) Panics(message string) {
	l.record(PANIC, message, true)
}

func (l *loggerModule) Panicf(format string, args ...any) {
	l.record(PANIC, fmt.Sprintf(format, args...), true)
}

func (l *loggerModule) record(category Category, message string, throwPanic bool) *Record {
	record := &Record{
		Category:  category,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}

	go l.write(record, throwPanic)

	return record
}

func (l *loggerModule) write(record *Record, throwPanic bool) *Record {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.records = append(l.records, *record)

	for _, m := range l.loggerModules {
		m.Vector(l.records)
		m.Record(record, throwPanic)
	}

	return record
}
