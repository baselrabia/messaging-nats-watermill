package watermill

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Connector interface {
	Subscribe(handlerName string, subscribeTopic string, handlerFunc message.NoPublishHandlerFunc) error
	Publish(topic string, messages ...*message.Message) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
