package main

import (
	"testing"
	"time"
)

func TestEventBusPublishWithNoSubscribers(t *testing.T) {
	eventBus := NewEventBus()

	eventBus.Publish(Event{
		Topic:     EventStreamChanged,
		StreamKey: "stream_key",
	})
}

func TestEventBusSubscribeReceivesMatchingEvent(t *testing.T) {
	eventBus := NewEventBus()
	subscription := eventBus.Subscribe(EventStreamChanged, StreamKeyFilter("stream_key"))
	defer subscription.Unregister()

	eventBus.Publish(Event{
		Topic:     EventStreamChanged,
		StreamKey: "stream_key",
	})

	waitForNotification(t, subscription.Notifications, 100*time.Millisecond)
}

func TestEventBusPublishSkipsNonMatchingFilter(t *testing.T) {
	eventBus := NewEventBus()
	subscription := eventBus.Subscribe(EventStreamChanged, StreamKeyFilter("stream_key"))
	defer subscription.Unregister()

	eventBus.Publish(Event{
		Topic:     EventStreamChanged,
		StreamKey: "other_stream_key",
	})

	assertNoNotification(t, subscription.Notifications, 50*time.Millisecond)
}

func TestEventBusUnregisterStopsNotifications(t *testing.T) {
	eventBus := NewEventBus()
	subscription := eventBus.Subscribe(EventStreamChanged, StreamKeyFilter("stream_key"))
	subscription.Unregister()

	eventBus.Publish(Event{
		Topic:     EventStreamChanged,
		StreamKey: "stream_key",
	})

	assertNoNotification(t, subscription.Notifications, 50*time.Millisecond)
}

func TestEventBusPublishNotifiesMultipleSubscribers(t *testing.T) {
	eventBus := NewEventBus()
	firstSubscription := eventBus.Subscribe(EventStreamChanged, StreamKeyFilter("stream_key"))
	secondSubscription := eventBus.Subscribe(EventStreamChanged, StreamKeyFilter("stream_key"))
	defer firstSubscription.Unregister()
	defer secondSubscription.Unregister()

	eventBus.Publish(Event{
		Topic:     EventStreamChanged,
		StreamKey: "stream_key",
	})

	waitForNotification(t, firstSubscription.Notifications, 100*time.Millisecond)
	waitForNotification(t, secondSubscription.Notifications, 100*time.Millisecond)
}

func TestEventBusPublishUsesTopicIsolation(t *testing.T) {
	eventBus := NewEventBus()
	subscription := eventBus.Subscribe(EventStreamChanged, StreamKeyFilter("stream_key"))
	defer subscription.Unregister()

	eventBus.Publish(Event{
		Topic:     EventTopic("other.topic"),
		StreamKey: "stream_key",
	})

	assertNoNotification(t, subscription.Notifications, 50*time.Millisecond)
}

func waitForNotification(t *testing.T, notifications <-chan struct{}, timeout time.Duration) {
	t.Helper()

	select {
	case <-notifications:
	case <-time.After(timeout):
		t.Fatal("timed out waiting for notification")
	}
}

func assertNoNotification(t *testing.T, notifications <-chan struct{}, duration time.Duration) {
	t.Helper()

	select {
	case <-notifications:
		t.Fatal("unexpected notification")
	case <-time.After(duration):
	}
}
