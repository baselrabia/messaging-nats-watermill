package nats

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// PublishMessage allows publishing a message to a specified subject.
func (w *WorkerImpl) PublishMessage(ctx context.Context, subject string, v interface{}) error {
	data, err := w.marshal(v)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal create message data")
	}

	msg := message.NewMessage(uuid.New().String(), data)
	
	err = w.connector.Publish(subject, msg)
	if err != nil {
		return errors.Wrap(err, "Failed to publish created msg")
	}

	return nil
}
