package nats

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	watermillNats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/baselrabia/messaging-nats-watermill/pkg/connectors/watermill/walrus"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Config - Defines the configuration needed to create a NATS connector.
type Config struct {
	NatsURL          string                         // URL to the NATS server
	NatsOptions      []nats.Option                  // NATS client options
	PublisherConfig  watermillNats.PublisherConfig  // Watermill NATS Publisher configuration
	SubscriberConfig watermillNats.SubscriberConfig // Watermill NATS Subscriber configuration
}

// Connector - Encapsulates the NATS client, Watermill router, and configuration.
type Connector struct {
	config     Config
	logger     watermill.LoggerAdapter
	client     *nats.Conn
	router     *message.Router
	publisher  message.Publisher
	subscriber message.Subscriber
}

// createClient - Initializes a NATS client with the provided configuration.
func createClient(config Config, log *logrus.Entry) (*nats.Conn, error) {
	// Connect to NATS server with options
	client, err := nats.Connect(config.NatsURL, config.NatsOptions...)
	if err != nil {
		log.WithError(err).Error("Failed to connect to NATS")
		return nil, err
	}
	return client, nil
}

// NewConnector - Creates a new NATS connector with a Watermill router and logger.
func NewConnector(config Config, log *logrus.Entry) (*Connector, error) {
	logger := walrus.NewWithLogger(log) // Use the Logrus-Watermill adapter

	// Initialize NATS client
	client, err := createClient(config, log)
	if err != nil {
		return nil, err
	}

	// Create a Watermill router
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create watermill router")
	}

	conn := &Connector{
		config: config,
		logger: logger,
		client: client,
		router: router,
	}

	return conn, nil
}

func (c *Connector) Publish(subject string, messages ...*message.Message) error {
	if c.publisher == nil {
		if err := c.initPublisher(); err != nil {
			return err
		}
	}
 
	if err := c.publisher.Publish(subject, messages...); err != nil {
		return errors.Wrap(err, "Failed to publish message.......")
	}

	return nil
}

func (c *Connector) initPublisher() error {
	publisher, err := watermillNats.NewPublisher(
		c.config.PublisherConfig,
		c.logger,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create NATS publisher")
	}
	c.publisher = publisher
	return nil
}

func (c *Connector) Subscribe(handlerName string, subject string, handlerFunc message.NoPublishHandlerFunc) error {
	if c.subscriber == nil {
		if err := c.initSubscriber(); err != nil {
			return err
		}
	}

	c.router.AddNoPublisherHandler(
		handlerName,
		subject,
		c.subscriber,
		handlerFunc,
	)

	 
	// // Start the router if not already started
	// go func() {
	// 	if err := c.router.Run(); err != nil {
	// 		c.logger.Error("Failed to run router", err, nil)
	// 	}
	// }()

	return nil
}

func (c *Connector) initSubscriber() error {
	// subscriber := gochannel.NewGoChannel(gochannel.Config{}, c.logger)

	subscriber, err := watermillNats.NewSubscriber(
		c.config.SubscriberConfig,
		c.logger,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create NATS subscriber")
	}
	c.subscriber = subscriber
	return nil
}

func (c *Connector) Start(ctx context.Context) error {
	if err := c.router.Run(ctx); err != nil {
		return errors.Wrap(err, "Failed to start router")
	}
	return nil
}

// Close gracefully shuts down the NATS connection and the Watermill router.
func (c *Connector) Stop(ctx context.Context) error {

	if err := c.router.Close(); err != nil {
		return errors.Wrap(err, "Failed to close router")
	}

	// if c.client != nil {
	// 	c.client.Close()
	// }

	// if c.publisher != nil {
	// 	if err := c.publisher.Close(); err != nil {
	// 		return errors.Wrap(err, "Failed to close publisher")
	// 	}
	// }

	// if c.subscriber != nil {
	// 	// Assuming the subscriber has a Close method, which may not be the case depending on your implementation.
	// 	// You might need to adjust this based on the actual interface of your NATS subscriber.
	// 	if err := c.subscriber.Close(); err != nil {
	// 		return errors.Wrap(err, "Failed to close subscriber")
	// 	}
	// }

	return nil
}

// // Create NATS Publisher
// publisher, err := watermillNats.NewPublisher(
// 	config.PublisherConfig,
// 	logger,
// )
// if err != nil {
// 	return nil, errors.Wrap(err, "Failed to create NATS publisher")
// }

// // Create NATS Subscriber
// subscriber, err := watermillNats.NewSubscriber(
// 	config.SubscriberConfig,
// 	logger,
// )
// if err != nil {
// 	return nil, errors.Wrap(err, "Failed to create NATS subscriber")
// }

// package nats

// import (
// 	"context"
// 	"github.com/ThreeDotsLabs/watermill"
// 	"github.com/ThreeDotsLabs/watermill/message"
// 	watermillNats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
// 	"github.com/nats-io/nats.go"
// 	"github.com/pkg/errors"
// 	"sync"
// 	"time"
// )

// // Ensure the SubscriberConfig struct correctly reflects NATS configurations
// type SubscriberConfig struct {
// 	URL              string
// 	QueueGroupPrefix string
// 	SubscribersCount int
// 	CloseTimeout     time.Duration
// 	AckWaitTimeout   time.Duration
// 	SubscribeTimeout time.Duration
// 	NatsOptions      []nats.Option
// 	Unmarshaler      watermillNats.Unmarshaler
// 	SubjectCalculator watermillNats.SubjectCalculator
// 	NakDelay         watermillNats.Delay
// 	JetStream        watermillNats.JetStreamConfig
// }

// type NatsConnector struct {
// 	publisher  message.Publisher
// 	subscriber *Subscriber
// }

// func NewNatsConnector(config SubscriberConfig, logger watermill.LoggerAdapter) (*NatsConnector, error) {
// 	if logger == nil {
// 		logger = watermill.NewStdLogger(false, false)
// 	}

// 	// Initialize NATS connection
// 	conn, err := nats.Connect(config.URL, config.NatsOptions...)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "cannot connect to NATS")
// 	}

// 	subscriberConfig := config.GetSubscriberSubscriptionConfig()
// 	subscriber, err := NewSubscriberWithNatsConn(conn, subscriberConfig, logger)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to create subscriber")
// 	}

// 	publisher, err := watermillNats.NewPublisher(
// 		watermillNats.PublisherConfig{
// 			URL:           config.URL,
// 			Marshaler:     watermillNats.NATSMarshaler{},
// 			NatsOptions:   config.NatsOptions,
// 			PublishOptions: nil, // Add any specific publish options here
// 		},
// 		logger,
// 	)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to create publisher")
// 	}

// 	return &NatsConnector{
// 		publisher:  publisher,
// 		subscriber: subscriber,
// 	}, nil
// }

// func (nc *NatsConnector) Publish(ctx context.Context, topic string, messages ...*message.Message) error {
// 	return nc.publisher.Publish(topic, messages...)
// }

// func (nc *NatsConnector) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
// 	return nc.subscriber.Subscribe(ctx, topic)
// }

// func (nc *NatsConnector) Close() error {
// 	if err := nc.publisher.Close(); err != nil {
// 		return err
// 	}
// 	return nc.subscriber.Close()
// }

// // package nats

// // import (
// // 	"github.com/ThreeDotsLabs/watermill"
// // 	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
// // 	"github.com/ThreeDotsLabs/watermill/message"
// // 	"log"
// // )

// // type Connector struct {
// // 	Publisher message.Publisher
// // 	Subscriber message.Subscriber
// // 	Logger watermill.LoggerAdapter
// // }

// // func NewConnector(natsURL string, logger watermill.LoggerAdapter) (*Connector, error) {
// // 	if logger == nil {
// // 		logger = watermill.NewStdLogger(false, false)
// // 	}

// // 	publisher, err := nats.NewPublisher(
// // 		nats.PublisherConfig{
// // 			URL: natsURL,
// // 		},
// // 		logger,
// // 	)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	subscriber, err := nats.NewSubscriber(
// // 		nats.SubscriberConfig{
// // 			URL:           natsURL,
// // 			QueueGroupPrefix: "watermill",
// // 		},
// // 		logger,
// // 	)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	return &Connector{
// // 		Publisher:  publisher,
// // 		Subscriber: subscriber,
// // 		Logger:     logger,
// // 	}, nil
// // }

// // func (c *Connector) Publish(subject string, msg *message.Message) error {
// // 	return c.Publisher.Publish(subject, msg)
// // }

// // func (c *Connector) Subscribe(subject string) (<-chan *message.Message, error) {
// // 	return c.Subscriber.Subscribe(subject)
// // }
