package subscription_test

import (
	"testing"

	"github.com/kubeshop/tracetest/server/subscription"
	"github.com/stretchr/testify/assert"
)

func TestManagerSubscriptionDifferentResources(t *testing.T) {
	manager := subscription.NewManager()
	var messageReceivedBySubscriber1, messageReceivedBySubscriber2 subscription.Message

	subscriber1 := subscription.NewSubscriberFunction(func(message subscription.Message) error {
		messageReceivedBySubscriber1 = message
		return nil
	})

	subscriber2 := subscription.NewSubscriberFunction(func(message subscription.Message) error {
		messageReceivedBySubscriber2 = message
		return nil
	})

	manager.Subscribe("test:1", subscriber1)
	manager.Subscribe("test:2", subscriber2)

	test1Message := subscription.Message{
		Type:    "test_update",
		Content: "test1 update",
	}

	test2Message := subscription.Message{
		Type:    "test_update",
		Content: "test2 update",
	}

	manager.PublishUpdate("test:1", test1Message)
	manager.PublishUpdate("test:2", test2Message)

	assert.Equal(t, test1Message, messageReceivedBySubscriber1, "received message should be equal to the one sent")
	assert.Equal(t, test2Message, messageReceivedBySubscriber2, "received message should be equal to the one sent")
}

func TestManagerSubscriptionSameResourceDifferentSubscribers(t *testing.T) {
	manager := subscription.NewManager()
	var messageReceivedBySubscriber1, messageReceivedBySubscriber2 subscription.Message

	subscriber1 := subscription.NewSubscriberFunction(func(message subscription.Message) error {
		messageReceivedBySubscriber1 = message
		return nil
	})

	subscriber2 := subscription.NewSubscriberFunction(func(message subscription.Message) error {
		messageReceivedBySubscriber2 = message
		return nil
	})

	manager.Subscribe("test:1", subscriber1)
	manager.Subscribe("test:1", subscriber2)

	test1Message := subscription.Message{
		Type:    "test_update",
		Content: "test1 update",
	}

	manager.PublishUpdate("test:1", test1Message)

	assert.NotNil(t, messageReceivedBySubscriber1, "message received by subscriber should not be nil")
	assert.Equal(t, messageReceivedBySubscriber1.Type, messageReceivedBySubscriber2.Type, "subscribers of the same resource should receive the same message")
	assert.Equal(t, messageReceivedBySubscriber1.Content, messageReceivedBySubscriber2.Content, "subscribers of the same resource should receive the same message")
}

func TestManagerUnsubscribe(t *testing.T) {
	manager := subscription.NewManager()
	var receivedMessage subscription.Message

	subscriber := subscription.NewSubscriberFunction(func(message subscription.Message) error {
		receivedMessage = message
		return nil
	})

	message1 := subscription.Message{
		Type:    "test_update",
		Content: "Test was updated",
	}

	message2 := subscription.Message{
		Type:    "test_deleted",
		Content: "Test was deleted",
	}

	manager.Subscribe("test:1", subscriber)
	manager.PublishUpdate("test:1", message1)

	assert.Equal(t, message1.Type, receivedMessage.Type)
	assert.Equal(t, message1.Content, receivedMessage.Content)

	manager.Unsubscribe("test:1", subscriber.ID())
	manager.PublishUpdate("test:1", message2)

	assert.Equal(t, message1.Type, receivedMessage.Type, "subscriber should not be notified after unsubscribed")
	assert.Equal(t, message1.Content, receivedMessage.Content, "subscriber should not be notified after unsubscribed")

}
