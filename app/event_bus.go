package main

import "sync"

type EventTopic string

const EventStreamChanged EventTopic = "stream.changed"

type Event struct {
	Topic     EventTopic
	StreamKey string
}

type EventFilter func(event Event) bool

func StreamKeyFilter(streamKey string) EventFilter {
	return func(event Event) bool {
		return event.StreamKey == streamKey
	}
}

type Subscription struct {
	Notifications <-chan struct{}
	Unregister    func()
}

type eventSubscriber struct {
	filter   EventFilter
	notifyCh chan struct{}
}

type EventBus struct {
	mutex       sync.Mutex
	nextID      uint64
	subscribers map[EventTopic]map[uint64]*eventSubscriber
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventTopic]map[uint64]*eventSubscriber),
	}
}

var (
	eventBusMutex    sync.RWMutex
	eventBusInstance = NewEventBus()
)

func GetEventBus() *EventBus {
	eventBusMutex.RLock()
	defer eventBusMutex.RUnlock()

	return eventBusInstance
}

func SetEventBusForTest(eventBus *EventBus) {
	eventBusMutex.Lock()
	defer eventBusMutex.Unlock()

	eventBusInstance = eventBus
}

func (eventBus *EventBus) Subscribe(topic EventTopic, filter EventFilter) Subscription {
	eventBus.mutex.Lock()
	defer eventBus.mutex.Unlock()

	if filter == nil {
		filter = func(event Event) bool {
			return true
		}
	}

	eventBus.nextID++
	subscriberID := eventBus.nextID
	notifyChannel := make(chan struct{}, 1)

	if eventBus.subscribers[topic] == nil {
		eventBus.subscribers[topic] = make(map[uint64]*eventSubscriber)
	}

	eventBus.subscribers[topic][subscriberID] = &eventSubscriber{
		filter:   filter,
		notifyCh: notifyChannel,
	}

	return Subscription{
		Notifications: notifyChannel,
		Unregister: func() {
			eventBus.mutex.Lock()
			defer eventBus.mutex.Unlock()

			delete(eventBus.subscribers[topic], subscriberID)
		},
	}
}

func (eventBus *EventBus) Publish(event Event) {
	eventBus.mutex.Lock()
	subscribers := eventBus.subscribers[event.Topic]
	subscriberSnapshot := make([]*eventSubscriber, 0, len(subscribers))
	for _, subscriber := range subscribers {
		subscriberSnapshot = append(subscriberSnapshot, subscriber)
	}
	eventBus.mutex.Unlock()

	for _, subscriber := range subscriberSnapshot {
		if !subscriber.filter(event) {
			continue
		}

		select {
		case subscriber.notifyCh <- struct{}{}:
		default:
		}
	}
}
