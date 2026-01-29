package system

import (
	"maps"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	str_topic "github.com/Rafael24595/go-api-core/src/commons/system/topic"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const SystemHubCategory = "SYSTEMHUB"

type SystemEvent struct {
	Topic string
	Value utils.Argument
}

func NewSystemEvent(topic, payload string) SystemEvent {
	return SystemEvent{
		Topic: topic,
		Value: *utils.ArgumentFrom(payload),
	}
}

type topic struct {
	parent    string
	status    bool
	timestamp int64
	channel   chan SystemEvent
}

type topics map[string]topic
type listeners map[string]topics

type TopicMeta struct {
	Parent    string
	Code      string
	Status    bool
	Timestamp int64
}

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

func (h *SystemEventHub) TopicsMeta(code string) []TopicMeta {
	l, ok := h.listeners[code]
	if !ok {
		return make([]TopicMeta, 0)
	}

	return collection.MapToDictionary(l, func(k string, v topic) TopicMeta {
		return TopicMeta{
			Code:      k,
			Parent:    v.parent,
			Status:    v.status,
			Timestamp: v.timestamp,
		}
	}).Values()
}

func (h *SystemEventHub) Subcribe(code string, listener chan SystemEvent, tcps ...str_topic.TopicAction) {
	if len(tcps) == 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	l, ok := h.listeners[code]
	if !ok {
		l = make(topics)
	}

	for _, t := range tcps {
		l[t.Code] = topic{
			parent:    t.Parent,
			status:    true,
			timestamp: time.Now().UnixMilli(),
			channel:   listener,
		}
	}

	h.listeners[code] = l
}

func (h *SystemEventHub) Unsubcribe(code string, topics ...str_topic.TopicAction) {
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
		delete(l, t.Parent)
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
			if t != event.Topic || !l.status {
				continue
			}

			select {
			case l.channel <- event:
			default:
				log.Customf(SystemHubCategory, "Dropped event for listener %q on topic %q: channel was full or not being read.", c, t)
			}
		}
	}
}
