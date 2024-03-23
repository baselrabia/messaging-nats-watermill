package main

import (
	"context"
	"fmt"
	"time"

	watermillNats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/baselrabia/messaging-nats-watermill/internals/workers/nats"

	wnc "github.com/baselrabia/messaging-nats-watermill/pkg/connectors/watermill/nats"
	nc "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

var (
	// For this example, we're using just a simple logger implementation,
	// You probably want to ship your own implementation of `watermill.LoggerAdapter`.
	logger    = logrus.NewEntry(logrus.StandardLogger())
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
	logger.Logger.SetLevel(logrus.DebugLevel)

	// Define the configuration for the NATS connection
	config := wnc.Config{
		NatsURL:     nc.DefaultURL,
		NatsOptions: options,
		PublisherConfig: watermillNats.PublisherConfig{
			URL:         natsURL,
			NatsOptions: options,
			Marshaler:   marshaler,
			JetStream:   jsConfig,
		},
		SubscriberConfig: watermillNats.SubscriberConfig{
			URL:            natsURL,
			CloseTimeout:   30 * time.Second,
			AckWaitTimeout: 30 * time.Second,
			NatsOptions:    options,
			Unmarshaler:    marshaler,
			JetStream:      jsConfig,
		},
	}

	// Initialize the connector here (assuming the existence of a suitable constructor)
	connector, err := wnc.NewConnector(config, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create connector")
	}

	// Create a new worker using the connector
	worker := nats.NewWorker(logger, nats.Config{NatsURL: config.NatsURL}, connector)

	// Define a message handler for the subscription
	handlerFunc := func(msg *message.Message) error {
		logger.Infof("Received message: %s", string(msg.Payload))
		// Acknowledge the message
		msg.Ack()
		return nil
	}

	// Setup subscription
	subject := "example.subject"
	handlerName := "example_handler"
	if err := worker.Consume(subject, handlerName,  handlerFunc); err != nil {
		logger.WithError(err).Fatal("Failed to consume messages")
	}

	go func() {
		err = connector.Start(context.Background())
		if err != nil {
			logger.WithError(err).Fatal("Failed to start connector")
		}
	}()

	// Simulate publishing a message
	testMsg := struct {
		Name string `json:"name"`
	}{Name: "Watermill"}

	logger.Info("Message subject  ", subject)

	// Publish a message every 10 seconds
	for {
		if err := worker.PublishMessage(context.Background(), subject, testMsg); err != nil {
			logger.WithError(err).Error("Failed to publish message")
		} else {
			logger.Info("Message published")
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
