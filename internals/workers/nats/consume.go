package nats

import (
	"github.com/ThreeDotsLabs/watermill/message"
)

// Consume sets up a subscription to a specified subject and processes incoming messages using a provided handler function.
func (w *WorkerImpl) Consume(subject, handlerName string, handlerFunc message.NoPublishHandlerFunc) error {
	// Subscribe using the connector with the provided handlerFunc.
	if err := w.connector.Subscribe(handlerName, subject, handlerFunc); err != nil {
		w.logger.WithError(err).Error("Failed to subscribe to subject")
		return err
	}

	return nil
}
