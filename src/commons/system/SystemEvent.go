package system

import (
	"maps"
	"sync"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const SystemHubCategory = "SYSTEMHUB"

type SystemEvent struct {
	Topic string
	Value utils.Argument
}

func NewSystemEvent(topic, payload string) SystemEvent {
	return  SystemEvent{
		Topic: topic,
		Value: *utils.ArgumentFrom(payload),
	}
}

type topic map[string]chan SystemEvent
type listeners map[string]topic

type SystemEventHub struct {
	mu        sync.Mutex
	listeners listeners
}

func InitializeSystemEventHub() *SystemEventHub {
	return &SystemEventHub{
		listeners: make(listeners),
	}
}

func (h *SystemEventHub) Topics(code string) []string {
	l, ok := h.listeners[code]
	if !ok {
		return make([]string, 0)
	}

	return collection.DictionaryFromMap(l).Keys()
}

func (h *SystemEventHub) Subcribe(code string, listener chan SystemEvent, topics ...string) {
	if len(topics) == 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	l, ok := h.listeners[code]
	if !ok {
		l = make(topic)
	}

	for _, t := range topics {
		l[t] = listener
	}

	h.listeners[code] = l
}

func (h *SystemEventHub) Unsubcribe(code string, topics ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(topics) == 0 {
		delete(h.listeners, code)
		return
	}

	l, ok := h.listeners[code]
	if !ok {
		return
	}

	for _, t := range topics {
		delete(l, t)
	}

	h.listeners[code] = l
}

func (h *SystemEventHub) Publish(topic, payload string) {
	event := SystemEvent{
		Topic: topic,
		Value: *utils.ArgumentFrom(payload),
	}

	go h.push(event)
}

func (h *SystemEventHub) push(event SystemEvent) {
	h.mu.Lock()
	snapshot := make(listeners, len(h.listeners))
	maps.Copy(snapshot, h.listeners)
	h.mu.Unlock()

	for c, ls := range snapshot {
		for t, l := range ls {
			if t != event.Topic {
				continue
			}

			select {
			case l <- event:
			default:
				log.Customf(SystemHubCategory, "Dropped event for listener %q on topic %q: channel was full or not being read.", c, t)
			}
		}
	}
}
