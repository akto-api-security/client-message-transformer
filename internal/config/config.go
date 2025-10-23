package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s", e.Message)
}

// Config holds all configuration from environment variables
type Config struct {
	SourceBrokers         string
	SourceTopic           string
	DestinationBrokers    string
	DestinationTopic      string
	ConsumerGroup         string
	LogLevel              string
	ClientID              string
	MaxConcurrentMessages int
	CommitInterval        time.Duration
	ProcessingTimeout     time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Required environment variables
	requiredVars := map[string]string{
		"CLIENT_ID":           os.Getenv("CLIENT_ID"),
		"SOURCE_BROKERS":      os.Getenv("SOURCE_BROKERS"),
		"SOURCE_TOPIC":        os.Getenv("SOURCE_TOPIC"),
		"DESTINATION_BROKERS": os.Getenv("DESTINATION_BROKERS"),
		"DESTINATION_TOPIC":   os.Getenv("DESTINATION_TOPIC"),
		"CONSUMER_GROUP":      os.Getenv("CONSUMER_GROUP"),
	}

	// Validate all required variables
	for varName, value := range requiredVars {
		if value == "" {
			return nil, &ConfigError{Message: fmt.Sprintf("%s environment variable is required but not configured", varName)}
		}
	}

	// Optional configuration with defaults
	config := &Config{
		SourceBrokers:         requiredVars["SOURCE_BROKERS"],
		SourceTopic:           requiredVars["SOURCE_TOPIC"],
		DestinationBrokers:    requiredVars["DESTINATION_BROKERS"],
		DestinationTopic:      requiredVars["DESTINATION_TOPIC"],
		ConsumerGroup:         requiredVars["CONSUMER_GROUP"],
		ClientID:              requiredVars["CLIENT_ID"],
		LogLevel:              getEnv("LOG_LEVEL", "INFO"),
		MaxConcurrentMessages: 10,
		CommitInterval:        5 * time.Second,
		ProcessingTimeout:     10 * time.Second,
	}

	return config, nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
