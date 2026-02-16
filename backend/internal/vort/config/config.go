// =============================================================================
// Package: config
// File: config.go
// Purpose: Configuration management for the VORT agent system
// Created: 2026-02-15
// =============================================================================
// This package provides configuration management for the VORT agent backend.
// It supports environment variable-based configuration with sensible defaults,
// runtime configuration updates, and configuration watching capabilities.
// =============================================================================

package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// =============================================================================
// Type: Config
// Purpose: Main configuration container for VORT agent system
// =============================================================================
// Config holds all configuration settings for the VORT agent system.
// It is thread-safe for reading and uses a mutex for updates.
//
// Usage:
//   cfg, err := config.Load()
//   addr := cfg.GetServerAddress()
type Config struct {
	mu sync.RWMutex

	// Logger instance
	Logger zerolog.Logger

	// Server configuration
	Server ServerConfig

	// Database connection settings
	Database DatabaseConfig

	// Message queue settings
	MessageQueue MessageQueueConfig

	// Agent-specific settings
	Agent AgentConfig

	// Encryption settings
	Encryption EncryptionConfig

	// Logging configuration
	Logging LoggingConfig
}

// =============================================================================
// Type: ServerConfig
// Purpose: HTTP server configuration
// =============================================================================
type ServerConfig struct {
	Host string // Server bind address (default: "0.0.0.0")
	Port int    // Server port (default: 8081)
}

// =============================================================================
// Type: DatabaseConfig
// Purpose: Database connection pool configuration
// =============================================================================
type DatabaseConfig struct {
	ConnectionString string        // PostgreSQL connection string
	MaxPoolSize      int          // Maximum connections in pool
	MinPoolSize      int          // Minimum idle connections
	MaxConnLifetime  time.Duration // Connection maximum lifetime
}

// =============================================================================
// Type: MessageQueueConfig
// Purpose: Message queue (RabbitMQ) connection settings
// =============================================================================
type MessageQueueConfig struct {
	Enabled        bool   // Enable/disable message queue
	Host           string // RabbitMQ host
	Port           int    // RabbitMQ port
	Username       string // Authentication username
	Password       string // Authentication password
	Exchange       string // Exchange name
	QueuePrefix    string // Queue name prefix for agents
}

// =============================================================================
// Type: AgentConfig
// Purpose: Agent behavior and operational settings
// =============================================================================
type AgentConfig struct {
	HeartbeatInterval    time.Duration // Agent heartbeat frequency
	CommandPollInterval  time.Duration // Command fetch interval
	MaxRetries           int          // Maximum command retry attempts
	RetryBackoff         time.Duration // Time between retry attempts
	MaxConcurrentCommands int         // Commands per agent simultaneously
	DataBufferSize       int          // In-memory data buffer size
}

// =============================================================================
// Type: EncryptionConfig
// Purpose: Data encryption and key management settings
// =============================================================================
type EncryptionConfig struct {
	Enabled         bool   // Enable data encryption
	KeyRotationDays int    // Days between key rotations
	Algorithm       string // Encryption algorithm (AES-256-GCM)
}

// =============================================================================
// Type: LoggingConfig
// Purpose: Logging output and format configuration
// =============================================================================
type LoggingConfig struct {
	Level      string // Log level (debug, info, warn, error)
	Format     string // Output format (json, text)
	OutputPath string // Output destination (stdout, file path)
}

// =============================================================================
// Function: Load
// Purpose: Load configuration from environment variables
// =============================================================================
// Load reads configuration from environment variables with fallback defaults.
// This function should be called at application startup.
//
// Environment Variables:
//   - VORT_SERVER_HOST: Server bind address
//   - VORT_SERVER_PORT: Server port
//   - VORT_DB_CONNECTION_STRING: PostgreSQL connection string
//   - VORT_DB_MAX_POOL_SIZE: Maximum database connections
//   - VORT_DB_MIN_POOL_SIZE: Minimum idle connections
//   - VORT_MQ_ENABLED: Enable message queue
//   - VORT_MQ_HOST: RabbitMQ host
//   - VORT_AGENT_HEARTBEAT_INTERVAL: Heartbeat frequency
//   - VORT_ENCRYPTION_ENABLED: Enable encryption
//   - VORT_LOG_LEVEL: Logging level
//
// Returns:
//   - *Config: Populated configuration
//   - error: Any configuration error
func Load() (*Config, error) {
	cfg := &Config{}

	// Logging defaults
	logLevel := getEnv("VORT_LOG_LEVEL", "info")
	logFormat := getEnv("VORT_LOG_FORMAT", "json")
	outputPath := getEnv("VORT_LOG_OUTPUT_PATH", "stdout")

	var level zerolog.Level
	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}

	if outputPath == "stdout" {
		logger := zerolog.New(os.Stdout).Level(level)
		if logFormat == "text" {
			logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		}
		cfg.Logger = logger
	} else {
		f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logger := zerolog.New(os.Stdout).Level(level)
			logger.Error().Err(err).Str("path", outputPath).Msg("failed to open log file, using stdout")
		} else {
			logger := zerolog.New(f).Level(level)
			if logFormat == "text" {
				logger = logger.Output(zerolog.ConsoleWriter{Out: f})
			}
			cfg.Logger = logger
		}
	}

	// Server defaults
	cfg.Server.Host = getEnv("VORT_SERVER_HOST", "0.0.0.0")
	cfg.Server.Port = getEnvInt("VORT_SERVER_PORT", 8081)

	// Database defaults
	cfg.Database.ConnectionString = getEnv("VORT_DB_CONNECTION_STRING",
		"postgres://postgres:postgres@localhost:5432/vortyx?sslmode=disable")
	cfg.Database.MaxPoolSize = getEnvInt("VORT_DB_MAX_POOL_SIZE", 25)
	cfg.Database.MinPoolSize = getEnvInt("VORT_DB_MIN_POOL_SIZE", 5)
	cfg.Database.MaxConnLifetime = getEnvDuration("VORT_DB_MAX_CONN_LIFETIME", time.Hour)

	// Message queue defaults
	cfg.MessageQueue.Enabled = getEnvBool("VORT_MQ_ENABLED", false)
	cfg.MessageQueue.Host = getEnv("VORT_MQ_HOST", "localhost")
	cfg.MessageQueue.Port = getEnvInt("VORT_MQ_PORT", 5672)
	cfg.MessageQueue.Username = getEnv("VORT_MQ_USERNAME", "guest")
	cfg.MessageQueue.Password = getEnv("VORT_MQ_PASSWORD", "guest")
	cfg.MessageQueue.Exchange = getEnv("VORT_MQ_EXCHANGE", "vortyx")
	cfg.MessageQueue.QueuePrefix = getEnv("VORT_MQ_QUEUE_PREFIX", "vort.agent.")

	// Agent defaults
	cfg.Agent.HeartbeatInterval = getEnvDuration("VORT_AGENT_HEARTBEAT_INTERVAL", 30*time.Second)
	cfg.Agent.CommandPollInterval = getEnvDuration("VORT_AGENT_COMMAND_POLL_INTERVAL", 5*time.Second)
	cfg.Agent.MaxRetries = getEnvInt("VORT_AGENT_MAX_RETRIES", 3)
	cfg.Agent.RetryBackoff = getEnvDuration("VORT_AGENT_RETRY_BACKOFF", time.Second)
	cfg.Agent.MaxConcurrentCommands = getEnvInt("VORT_AGENT_MAX_CONCURRENT_COMMANDS", 10)
	cfg.Agent.DataBufferSize = getEnvInt("VORT_AGENT_DATA_BUFFER_SIZE", 1000)

	// Encryption defaults
	cfg.Encryption.Enabled = getEnvBool("VORT_ENCRYPTION_ENABLED", true)
	cfg.Encryption.KeyRotationDays = getEnvInt("VORT_ENCRYPTION_KEY_ROTATION_DAYS", 30)
	cfg.Encryption.Algorithm = getEnv("VORT_ENCRYPTION_ALGORITHM", "AES-256-GCM")

	// Logging defaults
	cfg.Logging.Level = getEnv("VORT_LOG_LEVEL", "info")
	cfg.Logging.Format = getEnv("VORT_LOG_FORMAT", "json")
	cfg.Logging.OutputPath = getEnv("VORT_LOG_OUTPUT_PATH", "stdout")

	return cfg, nil
}

// GetServerAddress returns the server's bind address in host:port format.
//
// Returns:
//   - string: Formatted server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetAgentHeartbeatInterval returns the configured heartbeat interval.
// Thread-safe getter.
//
// Returns:
//   - time.Duration: Heartbeat interval
func (c *Config) GetAgentHeartbeatInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Agent.HeartbeatInterval
}

// GetCommandPollInterval returns the configured command poll interval.
// Thread-safe getter.
//
// Returns:
//   - time.Duration: Poll interval
func (c *Config) GetCommandPollInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Agent.CommandPollInterval
}

// UpdateAgentConfig applies runtime configuration updates to agent settings.
// This allows dynamic adjustment of agent behavior without restart.
//
// Parameters:
//   - ctx: Context for the update operation
//   - updates: Map of configuration keys to new values
//
// Supported keys:
//   - "heartbeat_interval": Duration string (e.g., "30s")
//   - "command_poll_interval": Duration string
//   - "max_retries": Integer
//
// Returns:
//   - error: Any update error
func (c *Config) UpdateAgentConfig(ctx context.Context, updates map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if interval, ok := updates["heartbeat_interval"].(string); ok {
		if d, err := time.ParseDuration(interval); err == nil {
			c.Agent.HeartbeatInterval = d
		}
	}

	if interval, ok := updates["command_poll_interval"].(string); ok {
		if d, err := time.ParseDuration(interval); err == nil {
			c.Agent.CommandPollInterval = d
		}
	}

	if maxRetries, ok := updates["max_retries"].(int); ok {
		c.Agent.MaxRetries = maxRetries
	}

	return nil
}

// =============================================================================
// Helper Functions
// Purpose: Environment variable parsing utilities
// =============================================================================

// getEnv retrieves an environment variable with a default fallback.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt parses an environment variable as integer.
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool parses an environment variable as boolean.
func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// getEnvDuration parses an environment variable as duration.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// =============================================================================
// Type: ConfigManager
// Purpose: Runtime configuration management and watching
// =============================================================================
// ConfigManager provides utilities for managing configuration at runtime,
// including hot-reloading and configuration watching.
//
// TODO: Add file-based configuration watching
// TODO: Add configuration validation before applying
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new configuration manager.
//
// Parameters:
//   - cfg: Initial configuration
//
// Returns:
//   - *ConfigManager: Manager instance
func NewConfigManager(cfg *Config) *ConfigManager {
	return &ConfigManager{config: cfg}
}

// GetConfig returns the current configuration.
//
// Returns:
//   - *Config: Current configuration
func (m *ConfigManager) GetConfig() *Config {
	return m.config
}

// Reload refreshes configuration from environment variables.
// This replaces the current configuration with newly loaded values.
//
// Returns:
//   - error: Any reload error
func (m *ConfigManager) Reload() error {
	newCfg, err := Load()
	if err != nil {
		return err
	}

	m.config.mu.Lock()
	defer m.config.mu.Unlock()
	*m.config = *newCfg

	return nil
}

// WatchConfig periodically reloads configuration in the background.
// This enables hot configuration updates without restart.
//
// Parameters:
//   - ctx: Context for cancellation
//   - reloadInterval: How often to check for updates
func (m *ConfigManager) WatchConfig(ctx context.Context, reloadInterval time.Duration) {
	ticker := time.NewTicker(reloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.Reload(); err != nil {
				fmt.Printf("Error reloading config: %v\n", err)
			}
		}
	}
}
