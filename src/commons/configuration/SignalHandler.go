package configuration

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Rafael24595/go-api-core/src/commons/log"
)

type signalHandler struct {
	sigChan chan os.Signal
	done    chan bool
}

func newSignalHandler() *signalHandler {
	h := &signalHandler{
		sigChan: make(chan os.Signal, 1),
		done:    make(chan bool, 1),
	}

	signal.Notify(h.sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-h.sigChan
		log.Message("Shutdown signal received.")
		h.done <- true
	}()

	return h
}

func (h *signalHandler) Done() <-chan bool {
	return h.done
}
