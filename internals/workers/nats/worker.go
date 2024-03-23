package nats

import (
	"context"
	"encoding/json"

 	"github.com/baselrabia/messaging-nats-watermill/pkg/connectors/watermill"
	"github.com/sirupsen/logrus"
)

// Worker interface as previously defined
type Worker interface {
	PublishMessage(ctx context.Context, subject string, v interface{}) error
}

// Config struct as previously defined
type Config struct {
	NatsURL string
	// Add other configuration fields as necessary.
}

// WorkerImpl struct adjusted
type WorkerImpl struct {
	logger    *logrus.Entry
	config    Config
	connector watermill.Connector // Assuming this is a type that satisfies some interface requirements.
	marshal   func(v interface{}) ([]byte, error)
	unmarshal func(data []byte, v interface{}) error
}

// NewWorker function adjusted based on the new requirement
func NewWorker(
	logger *logrus.Entry,
	config Config,
	connector watermill.Connector,
) *WorkerImpl {
	return &WorkerImpl{
		logger:    logger,
		config:    config,
		connector: connector,
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}
}
