package system_test

import (
	"sync"
	"testing"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/test/support/assert"
)

func TestSubscribeAndPublish(t *testing.T) {
	hub := system.InitializeSystemEventHub()
	ch := make(chan system.SystemEvent, 1)

	hub.Subcribe("listener_1", ch, "langs")
	hub.Publish("langs", "golang")

	select {
	case e := <-ch:
		assert.Equal(t, "langs", e.Topic)
		assert.Equal(t, "golang", e.Value.String())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestSubscribeAndPublish_MultipleListeners(t *testing.T) {
	hub := system.InitializeSystemEventHub()
	ch1 := make(chan system.SystemEvent, 1)
	ch2 := make(chan system.SystemEvent, 1)

	hub.Subcribe("listener_1", ch1, "langs")
	hub.Subcribe("listener_2", ch2, "langs")

	hub.Publish("langs", "zig")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case e := <-ch1:
			assert.Equal(t, "langs", e.Topic)
			assert.Equal(t, "zig", e.Value.String())
		case <-time.After(100 * time.Millisecond):
			t.Error("listener 1 did not receive message")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case e := <-ch2:
			assert.Equal(t, "langs", e.Topic)
			assert.Equal(t, "zig", e.Value.String())
		case <-time.After(100 * time.Millisecond):
			t.Error("listener 2 did not receive message")
		}
	}()

	wg.Wait()
}

func TestUnsubscribe_SingleTopic(t *testing.T) {
	hub := system.InitializeSystemEventHub()
	ch := make(chan system.SystemEvent, 1)

	hub.Subcribe("listener_1", ch, "langs", "dbs")

	hub.Unsubcribe("listener_1", "langs")

	hub.Publish("langs", "rust")
	hub.Publish("dbs", "surrealdb")

	select {
	case e := <-ch:
		assert.Equal(t, "dbs", e.Topic)
		assert.Equal(t, "surrealdb", e.Value.String())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected message on topic2")
	}
}

func TestUnsubscribe_AllTopics(t *testing.T) {
	hub := system.InitializeSystemEventHub()
	ch := make(chan system.SystemEvent, 1)

	hub.Subcribe("listener_1", ch, "langs", "dbs")

	hub.Unsubcribe("listener_1")

	hub.Publish("langs", "rust")
	hub.Publish("dbs", "surrealdb")

	select {
	case <-ch:
		t.Fatal("expected no message after full unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// OK
	}
}

func TestSubscribeAndPublish_NonSubscribedTopic(t *testing.T) {
	hub := system.InitializeSystemEventHub()
	ch := make(chan system.SystemEvent, 1)

	hub.Subcribe("listener_1", ch, "langs")
	hub.Publish("postgresql", "surrealdb")

	select {
	case <-ch:
		t.Fatal("received event for non-subscribed topic")
	case <-time.After(50 * time.Millisecond):
		// OK
	}
}

func TestSubscribeAndPublish_DroppedEventLogsWarning(t *testing.T) {
	hub := system.InitializeSystemEventHub()
	ch := make(chan system.SystemEvent, 1)

	hub.Subcribe("listener_1", ch, "langs")

	ch <- system.NewSystemEvent("langs", "elixir")

	hub.Publish("langs", "clojure")

	time.Sleep(100 * time.Millisecond)

	records := log.Records()
	if len(records) < 2 {
		t.Fatal("the error is not logged")
	}

	assert.Equal(t, system.SystemHubCategory, records[1].Category)

	time.Sleep(100 * time.Millisecond)

	select {
	case <-ch:
	default:
	}
}

func TestSubscribeAndPublish_Concurrent(t *testing.T) {
	hub := system.InitializeSystemEventHub()
	ch := make(chan system.SystemEvent, 10)
	hub.Subcribe("listener_1", ch, "langs")

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			hub.Publish("langs", "golang")
		}(i)
	}
	wg.Wait()

	count := 0
L:
	for {
		select {
		case <-ch:
			count++
		case <-time.After(500 * time.Millisecond):
			break L
		}
	}

	assert.GreaterOrEqual(t, 1, count, "expected at least one message")
}
