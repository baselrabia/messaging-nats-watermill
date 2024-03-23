// Sources for https://watermill.io/docs/getting-started/
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watermillNats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	nc "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

var (
	// For this example, we're using just a simple logger implementation,
	// You probably want to ship your own implementation of `watermill.LoggerAdapter`.
	logger    = watermill.NewStdLogger(false, false)
	natsURL   = nc.DefaultURL
	marshaler = &watermillNats.GobMarshaler{}
	options   = []nc.Option{
		nc.RetryOnFailedConnect(true),
		nc.Timeout(30 * time.Second),
		nc.ReconnectWait(1 * time.Second),
	}
	jsConfig = watermillNats.JetStreamConfig{Disabled: true}
)

func main() {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	// SignalsHandler will gracefully shutdown Router when SIGTERM is received.
	// You can also close the router by just calling `r.Close()`.
	router.AddPlugin(plugin.SignalsHandler)

	// For simplicity, we are using the gochannel Pub/Sub here,
	// You can replace it with any Pub/Sub implementation, it will work the same.
	_ = gochannel.NewGoChannel(gochannel.Config{}, logger)

	subscriber, err := watermillNats.NewSubscriber(
		watermillNats.SubscriberConfig{
			URL:            natsURL,
			CloseTimeout:   30 * time.Second,
			AckWaitTimeout: 30 * time.Second,
			NatsOptions:    options,
			Unmarshaler:    marshaler,
			JetStream:      jsConfig,
		},
		logger,
	)
	if err != nil {
		panic(errors.Wrap(err, "Failed to create NATS subscriber"))
	}
	publisher, err := watermillNats.NewPublisher(
				watermillNats.PublisherConfig{
					URL:         natsURL,
					NatsOptions: options,
					Marshaler:   marshaler,
					JetStream:   jsConfig,
				},
				logger,
			)
			if err != nil {
				panic(err)
			}

		
	// Producing some incoming messages in background
	go publishMessages(publisher)

	// just for debug, we are printing all messages received on `incoming_messages_topic`
	router.AddNoPublisherHandler(
		"print_incoming_messages",
		"incoming_messages_topic",
		subscriber,
		printMessages,
	)

	// Now that all handlers are registered, we're running the Router.
	// Run is blocking while the router is running.
	ctx := context.Background()
	if err := router.Run(ctx); err != nil {
		panic(err)
	}
}

func publishMessages(publisher message.Publisher) {
	for {
		time.Sleep(3 * time.Second)
		msg := message.NewMessage(watermill.NewUUID(), []byte("Hello, world!"))
		middleware.SetCorrelationID(watermill.NewUUID(), msg)

		log.Printf("sending message %s, correlation id: %s\n", msg.UUID, middleware.MessageCorrelationID(msg))

		if err := publisher.Publish("incoming_messages_topic", msg); err != nil {
			panic(err)
		}

		time.Sleep(10 * time.Second)
	}
}

func printMessages(msg *message.Message) error {
	fmt.Printf(
		"\n> Received message: %s\n> %s\n> metadata: %v\n\n",
		msg.UUID, string(msg.Payload), msg.Metadata,
	)
	return nil
}

type structHandler struct {
	// we can add some dependencies here
}

func (s structHandler) Handler(msg *message.Message) ([]*message.Message, error) {
	log.Println("structHandler received message", msg.UUID)

	msg = message.NewMessage(watermill.NewUUID(), []byte("message produced by structHandler"))
	return message.Messages{msg}, nil
}
