package local

import (
	"fmt"

	"github.com/Rafael24595/go-log/log"
)

// TODO: Remove.

func Panic(err error) {
	log.Error(err)
	panic(err)
}

func Panics(message string) {
	log.Message(message)
	panic(message)
}

func Panicf(format string, args ...any) {
	log.Messagef(format, args...)
	panic(fmt.Sprintf(format, args...))
}
