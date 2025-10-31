package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// ClientConfig holds Kafka client configuration
type ClientConfig struct {
	Brokers          string
	ConsumerGroup    string
	Topic            string
	SASLEnabled      bool
	SASLMechanism    string
	SASLUsername     string
	SASLPassword     string
	SecurityProtocol string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config *ClientConfig) (*kafka.Consumer, error) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers":               config.Brokers,
		"group.id":                        config.ConsumerGroup,
		"auto.offset.reset":               "earliest",
		"enable.auto.commit":              false,
		"go.application.rebalance.enable": true,
		"socket.keepalive.enable":         true,
		"socket.timeout.ms":               60000,
		"api.version.request.timeout.ms":  30000,
		"reconnect.backoff.ms":            100,
		"reconnect.backoff.max.ms":        10000,
		"metadata.max.age.ms":             300000,
	}

	// Add SASL configuration if enabled
	if config.SASLEnabled {
		configMap.SetKey("security.protocol", config.SecurityProtocol)
		configMap.SetKey("sasl.mechanism", config.SASLMechanism)
		configMap.SetKey("sasl.username", config.SASLUsername)
		configMap.SetKey("sasl.password", config.SASLPassword)
		fmt.Printf("🔐 Consumer SASL Config: protocol=%s, mechanism=%s, username=%s\n",
			config.SecurityProtocol, config.SASLMechanism, config.SASLUsername)
	} else {
		fmt.Printf("⚠️  Consumer SASL DISABLED\n")
	}

	consumer, err := kafka.NewConsumer(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return consumer, nil
}

// NewProducer creates a new Kafka producer with retry logic
func NewProducer(config *ClientConfig) (*kafka.Producer, error) {
	maxRetries := 5
	retryDelay := time.Second * 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		configMap := &kafka.ConfigMap{
			"bootstrap.servers":                     config.Brokers,
			"acks":                                  "all",
			"retries":                               10,
			"max.in.flight.requests.per.connection": 5,
			"socket.keepalive.enable":               true,
			"socket.timeout.ms":                     60000,
			"api.version.request.timeout.ms":        30000,
			"reconnect.backoff.ms":                  100,
			"reconnect.backoff.max.ms":              10000,
			"metadata.max.age.ms":                   300000,
			"delivery.timeout.ms":                   300000,
		}

		// Add SASL configuration if enabled
		if config.SASLEnabled {
			configMap.SetKey("security.protocol", config.SecurityProtocol)
			configMap.SetKey("sasl.mechanism", config.SASLMechanism)
			configMap.SetKey("sasl.username", config.SASLUsername)
			configMap.SetKey("sasl.password", config.SASLPassword)
			fmt.Printf("🔐 Producer SASL Config: protocol=%s, mechanism=%s, username=%s\n",
				config.SecurityProtocol, config.SASLMechanism, config.SASLUsername)
		} else {
			fmt.Printf("⚠️  Producer SASL DISABLED\n")
		}

		producer, err := kafka.NewProducer(configMap)
		if err == nil {
			fmt.Printf("✅ Producer connected to %s\n", config.Brokers)
			return producer, nil
		}

		if attempt < maxRetries {
			fmt.Printf("⏳ Producer connection attempt %d/%d failed, retrying in %v...\n", attempt, maxRetries, retryDelay)
			time.Sleep(retryDelay)
			retryDelay = time.Duration(float64(retryDelay) * 1.5) // Exponential backoff with 1.5x multiplier
		} else {
			return nil, fmt.Errorf("failed to create producer after %d attempts: %w", maxRetries, err)
		}
	}

	return nil, fmt.Errorf("failed to create producer")
}
