package log

import (
	"fmt"
	"sync"
	"time"
)

const CODE_CONSOLE = "CONSOLE"

type console struct {
	mu        sync.Mutex
	formatter Formatter
	records   []Record
}

func newConsole(formatter *Formatter) *console {
	return &console{
		formatter: *formatter,
		records:   make([]Record, 0),
	}
}

func (l *console) Name() string {
	return CODE_CONSOLE
}

func (l *console) Records() []Record {
	l.mu.Lock()
	defer l.mu.Unlock()

	clone := make([]Record, len(l.records))
    copy(clone, l.records)

	return clone
}

func (l *console) Message(message string) *Record {
	return l.record(MESSAGE, message, false)
}

func (l *console) Messagef(format string, args ...any) *Record {
	return l.record(MESSAGE, fmt.Sprintf(format, args...), false)
}

func (l *console) Warning(message string) *Record {
	return l.record(WARNING, message, false)
}

func (l *console) Warningf(format string, args ...any) *Record {
	return l.record(WARNING, fmt.Sprintf(format, args...), false)
}

func (l *console) Error(err error) *Record {
	return l.record(ERROR, err.Error(), false)
}

func (l *console) Errors(message string) *Record {
	return l.record(ERROR, message, false)
}

func (l *console) Errorf(format string, args ...any) *Record {
	return l.record(ERROR, fmt.Sprintf(format, args...), false)
}

func (l *console) Panic(err error) {
	l.record(PANIC, err.Error(), true)
}

func (l *console) Panics(message string) {
	l.record(PANIC, message, true)
}

func (l *console) Panicf(format string, args ...any) {
	l.record(PANIC, fmt.Sprintf(format, args...), true)
}

func (l *console) record(category Category, message string, throwPanic bool) *Record {
	record := &Record{
		Category: category,
		Message: message,
		Timestamp: time.Now().UnixMilli(),
	}

	go l.write(record, throwPanic)

	return record
}

func (l *console) write(record *Record, throwPanic bool) *Record {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.records = append(l.records, *record)

	message := l.formatter.Format(*record)

	println(message)

	if throwPanic {
		panic(message)
	}
	
	return record
}
