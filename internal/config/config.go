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

	// Source SASL Configuration
	SourceSASLEnabled      bool
	SourceSASLMechanism    string
	SourceSASLUsername     string
	SourceSASLPassword     string
	SourceSecurityProtocol string

	// Destination SASL Configuration
	DestinationSASLEnabled      bool
	DestinationSASLMechanism    string
	DestinationSASLUsername     string
	DestinationSASLPassword     string
	DestinationSecurityProtocol string
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

		// Source SASL Configuration (optional)
		SourceSASLEnabled:      getEnvBool("SOURCE_SASL_ENABLED", false),
		SourceSASLMechanism:    getEnv("SOURCE_SASL_MECHANISM", "PLAIN"),
		SourceSASLUsername:     getEnv("SOURCE_SASL_USERNAME", ""),
		SourceSASLPassword:     getEnv("SOURCE_SASL_PASSWORD", ""),
		SourceSecurityProtocol: getEnv("SOURCE_SECURITY_PROTOCOL", "SASL_PLAINTEXT"),

		// Destination SASL Configuration (optional)
		DestinationSASLEnabled:      getEnvBool("DESTINATION_SASL_ENABLED", false),
		DestinationSASLMechanism:    getEnv("DESTINATION_SASL_MECHANISM", "PLAIN"),
		DestinationSASLUsername:     getEnv("DESTINATION_SASL_USERNAME", ""),
		DestinationSASLPassword:     getEnv("DESTINATION_SASL_PASSWORD", ""),
		DestinationSecurityProtocol: getEnv("DESTINATION_SECURITY_PROTOCOL", "SASL_PLAINTEXT"),
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

// getEnvBool gets boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "TRUE" || value == "1"
	}
	return defaultValue
}
