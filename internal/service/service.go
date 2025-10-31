package service

import (
	"client-message-transformer/internal/config"
	"client-message-transformer/internal/kafka"
	"client-message-transformer/internal/logger"
	"client-message-transformer/internal/metrics"
	"client-message-transformer/internal/transformer"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// TransformerService handles message transformation
type TransformerService struct {
	config   *config.Config
	consumer *kafkalib.Consumer
	producer *kafkalib.Producer
	logger   *logger.Logger
	metrics  *metrics.Metrics
	stopChan chan bool
	wg       sync.WaitGroup
}

// New creates a new transformer service
func New(cfg *config.Config) (*TransformerService, error) {
	log := logger.NewLogger(cfg.LogLevel)

	log.Info("╔════════════════════════════════════════════════════════════╗")
	log.Info("║        Initializing Kafka Transformer Service             ║")
	log.Info("╚════════════════════════════════════════════════════════════╝")
	log.Info("")

	log.Info("📋 === SOURCE BROKER DETAILS ===")
	log.Info(fmt.Sprintf("   🔗 Bootstrap Servers: %s", cfg.SourceBrokers))
	log.Info(fmt.Sprintf("   📍 Topic: %s", cfg.SourceTopic))
	log.Info(fmt.Sprintf("   👥 Consumer Group: %s", cfg.ConsumerGroup))
	log.Info("")

	log.Info("📋 === DESTINATION BROKER DETAILS ===")
	log.Info(fmt.Sprintf("   🔗 Bootstrap Servers: %s", cfg.DestinationBrokers))
	log.Info(fmt.Sprintf("   📍 Topic: %s", cfg.DestinationTopic))
	log.Info("")

	log.Info("⏳ Waiting for Kafka brokers to be ready...")
	time.Sleep(5 * time.Second) // Give Kafka time to fully initialize

	// Create consumer
	consumerCfg := &kafka.ClientConfig{
		Brokers:          cfg.SourceBrokers,
		ConsumerGroup:    cfg.ConsumerGroup,
		Topic:            cfg.SourceTopic,
		SASLEnabled:      cfg.SourceSASLEnabled,
		SASLMechanism:    cfg.SourceSASLMechanism,
		SASLUsername:     cfg.SourceSASLUsername,
		SASLPassword:     cfg.SourceSASLPassword,
		SecurityProtocol: cfg.SourceSecurityProtocol,
	}
	log.Info(fmt.Sprintf("� Attempting to connect to source broker: %s", cfg.SourceBrokers))
	consumer, err := kafka.NewConsumer(consumerCfg)
	if err != nil {
		log.Error(fmt.Sprintf("❌ Failed to create consumer: %v", err))
		return nil, err
	}
	log.Info("✅ Consumer connected to source broker successfully")

	// Create producer
	log.Info(fmt.Sprintf("� Attempting to connect to destination broker: %s", cfg.DestinationBrokers))
	producerCfg := &kafka.ClientConfig{
		Brokers:          cfg.DestinationBrokers,
		SASLEnabled:      cfg.DestinationSASLEnabled,
		SASLMechanism:    cfg.DestinationSASLMechanism,
		SASLUsername:     cfg.DestinationSASLUsername,
		SASLPassword:     cfg.DestinationSASLPassword,
		SecurityProtocol: cfg.DestinationSecurityProtocol,
	}
	producer, err := kafka.NewProducer(producerCfg)
	if err != nil {
		log.Error(fmt.Sprintf("❌ Failed to create producer: %v", err))
		consumer.Close()
		return nil, err
	}
	log.Info("✅ Producer connected to destination broker successfully")

	service := &TransformerService{
		config:   cfg,
		consumer: consumer,
		producer: producer,
		logger:   log,
		metrics:  metrics.New(),
		stopChan: make(chan bool),
	}

	log.Info("")
	log.Info("╔════════════════════════════════════════════════════════════╗")
	log.Info("║           ✅ Service Initialized Successfully              ║")
	log.Info("╚════════════════════════════════════════════════════════════╝")
	log.Info("")
	log.Info("📥 SOURCE CONFIGURATION:")
	log.Info(fmt.Sprintf("   Broker: %s | Topic: %s | Group: %s", cfg.SourceBrokers, cfg.SourceTopic, cfg.ConsumerGroup))
	log.Info("")
	log.Info("📤 DESTINATION CONFIGURATION:")
	log.Info(fmt.Sprintf("   Broker: %s | Topic: %s", cfg.DestinationBrokers, cfg.DestinationTopic))
	log.Info("")
	log.Info("🚀 Ready to process messages...")
	log.Info("")

	return service, nil
}

// Start begins processing messages
func (s *TransformerService) Start(ctx context.Context) error {
	// Wait additional time for broker metadata to be fully loaded
	s.logger.Info("⏳ Waiting for broker metadata...")
	time.Sleep(3 * time.Second)

	err := s.consumer.SubscribeTopics([]string{s.config.SourceTopic}, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to subscribe: %v", err))
		return err
	}

	s.logger.Info(fmt.Sprintf("✅ Subscribed to topic: %s", s.config.SourceTopic))

	s.wg.Add(1)
	go s.processMessages(ctx)

	s.wg.Add(1)
	go s.reportMetrics(ctx)

	s.logger.Info("🚀 Message processing started")
	return nil
}

// processMessages main event loop
func (s *TransformerService) processMessages(ctx context.Context) {
	defer s.wg.Done()

	semaphore := make(chan bool, s.config.MaxConcurrentMessages)
	commitTicker := time.NewTicker(s.config.CommitInterval)
	defer commitTicker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("Shutting down message processing...")
			return

		case <-ctx.Done():
			s.logger.Info("Context cancelled")
			return

		case <-commitTicker.C:
			_, err := s.consumer.Commit()
			if err != nil && err.(kafkalib.Error).Code() != kafkalib.ErrNoOffset {
				s.logger.Warn(fmt.Sprintf("Commit failed: %v", err))
			}

		default:
			msg, err := s.consumer.ReadMessage(s.config.ProcessingTimeout)
			if err != nil {
				kafkaErr, ok := err.(kafkalib.Error)
				if ok && kafkaErr.Code() == kafkalib.ErrTimedOut {
					// Timeout is normal, just continue
					continue
				}
				s.logger.Error(fmt.Sprintf("Consumer error: %v (type: %T)", err, err))
				continue
			}

			// Message received!
			s.logger.Info(fmt.Sprintf("📨 Message received from topic %s (size: %d bytes)", s.config.SourceTopic, len(msg.Value)))
			s.logger.Debug(fmt.Sprintf("Message content: %s", string(msg.Value)))

			semaphore <- true
			s.wg.Add(1)

			go func(kafkaMsg *kafkalib.Message) {
				defer s.wg.Done()
				defer func() { <-semaphore }()
				s.processMessage(kafkaMsg)
			}(msg)
		}
	}
}

// processMessage transforms a single message
func (s *TransformerService) processMessage(kafkaMsg *kafkalib.Message) {
	startTime := time.Now()

	clientID := s.config.ClientID
	s.logger.Info(fmt.Sprintf("🔄 Processing message for client: %s", clientID))

	s.metrics.IncrementReceived()

	// Transform message
	s.logger.Debug(fmt.Sprintf("Raw message: %s", string(kafkaMsg.Value)))
	transformed, err := transformer.TransformMessage(kafkaMsg.Value, clientID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("❌ Transformation failed: %v", err))
		s.metrics.IncrementFailed()
		return
	}

	s.logger.Info("✅ Message transformed successfully")
	s.metrics.IncrementTransformed()

	// Marshal to JSON
	transformedJSON, err := json.Marshal(transformed)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal: %v", err))
		s.metrics.IncrementFailed()
		return
	}

	// Publish
	err = s.publishMessage(clientID, transformedJSON)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to publish: %v", err))
		s.metrics.IncrementFailed()
		return
	}

	s.metrics.IncrementPublished()
	s.metrics.AddProcessingTime(time.Since(startTime))

	s.logger.Debug(fmt.Sprintf("✅ Message processed in %v (client: %s)", time.Since(startTime), clientID))
}

// publishMessage sends transformed message to destination (non-blocking)
func (s *TransformerService) publishMessage(clientID string, data []byte) error {
	err := s.producer.Produce(
		&kafkalib.Message{
			TopicPartition: kafkalib.TopicPartition{
				Topic:     &s.config.DestinationTopic,
				Partition: kafkalib.PartitionAny,
			},
			Key:   []byte(clientID),
			Value: data,
			Headers: []kafkalib.Header{
				{Key: "client_id", Value: []byte(clientID)},
				{Key: "transformed_at", Value: []byte(time.Now().Format(time.RFC3339))},
			},
		},
		nil, // No delivery callback - non-blocking
	)

	if err != nil {
		return fmt.Errorf("failed to produce message to %s: %w", s.config.DestinationTopic, err)
	}

	// Flush to ensure message is queued
	remaining := s.producer.Flush(5000) // 5 second timeout
	if remaining > 0 {
		s.logger.Error(fmt.Sprintf("⚠️  Warning: %d messages remained in queue after flush", remaining))
	}

	s.logger.Info(fmt.Sprintf("📤 Published to %s (client: %s)", s.config.DestinationTopic, clientID))
	return nil
}

// extractClientID extracts client ID from message
func (s *TransformerService) extractClientID(kafkaMsg *kafkalib.Message) string {
	// Try headers
	for _, header := range kafkaMsg.Headers {
		if header.Key == "client_id" {
			return string(header.Value)
		}
	}

	// Try payload
	var data map[string]interface{}
	if err := json.Unmarshal(kafkaMsg.Value, &data); err == nil {
		if clientID, ok := data["akto_account_id"].(string); ok && clientID != "" {
			return clientID
		}
	}

	return "default-client"
}

// reportMetrics logs metrics periodically
func (s *TransformerService) reportMetrics(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(60 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.printMetrics()
		}
	}
}

// printMetrics logs current metrics
func (s *TransformerService) printMetrics() {
	snapshot := s.metrics.GetSnapshot()

	s.logger.Info("📊 === METRICS REPORT ===")
	s.logger.Info(fmt.Sprintf("   Received:    %d messages", snapshot["received"].(int64)))
	s.logger.Info(fmt.Sprintf("   Transformed: %d messages", snapshot["transformed"].(int64)))
	s.logger.Info(fmt.Sprintf("   Published:   %d messages", snapshot["published"].(int64)))
	s.logger.Info(fmt.Sprintf("   Failed:      %d messages", snapshot["failed"].(int64)))
	s.logger.Info(fmt.Sprintf("   Avg Time:    %v", snapshot["avg_time"].(time.Duration)))
	s.logger.Info("📊 ========================")
}

// Stop gracefully shuts down the service
func (s *TransformerService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping service...")

	close(s.stopChan)

	done := make(chan bool)
	go func() {
		s.wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		s.logger.Info("✅ All goroutines stopped")
	case <-ctx.Done():
		s.logger.Warn("⚠️ Shutdown timeout exceeded")
	}

	s.consumer.Close()
	s.producer.Close()

	s.logger.Info("✅ Service stopped")
	s.printMetrics()
	return nil
}
