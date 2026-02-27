package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Azure    AzureConfig
	Logging  LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            string
	Environment     string
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// AzureConfig holds Azure service configuration
type AzureConfig struct {
	OpenAI  OpenAIConfig
	Speech  SpeechConfig
	Storage StorageConfig
}

// OpenAIConfig holds Azure OpenAI configuration
type OpenAIConfig struct {
	Endpoint   string
	APIKey     string
	Deployment string
}

// SpeechConfig holds Azure Speech Service configuration
type SpeechConfig struct {
	SubscriptionKey string
	Region          string
	Endpoint        string
}

// StorageConfig holds Azure Blob Storage configuration
type StorageConfig struct {
	AccountName      string
	AccountKey       string
	ConnectionString string
	BlobEndpoint     string
	AudioContainer   string
	ReportContainer  string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string // json or console
}

// Load reads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from environment variables
	v.AutomaticEnv()

	// Bind specific environment variables
	bindEnvVars(v)

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.environment", "development")
	v.SetDefault("server.shutdowntimeout", 30*time.Second)

	// Database defaults
	v.SetDefault("database.maxopenconns", 25)
	v.SetDefault("database.maxidleconns", 5)
	v.SetDefault("database.connmaxlifetime", 5*time.Minute)

	// Azure Storage defaults
	v.SetDefault("azure.storage.audiocontainer", "audio-recordings")
	v.SetDefault("azure.storage.reportcontainer", "health-reports")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

// bindEnvVars binds environment variables to config keys
func bindEnvVars(v *viper.Viper) {
	// Server
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.environment", "ENV", "ENVIRONMENT")

	// Database
	v.BindEnv("database.url", "DATABASE_URL")

	// Azure OpenAI
	v.BindEnv("azure.openai.endpoint", "AZURE_OPENAI_ENDPOINT")
	v.BindEnv("azure.openai.apikey", "AZURE_OPENAI_API_KEY")
	v.BindEnv("azure.openai.deployment", "AZURE_OPENAI_DEPLOYMENT")

	// Azure Speech
	v.BindEnv("azure.speech.subscriptionkey", "AZURE_SPEECH_KEY")
	v.BindEnv("azure.speech.region", "AZURE_SPEECH_REGION")
	v.BindEnv("azure.speech.endpoint", "AZURE_SPEECH_ENDPOINT")

	// Azure Storage
	v.BindEnv("azure.storage.accountname", "AZURE_STORAGE_ACCOUNT_NAME")
	v.BindEnv("azure.storage.accountkey", "AZURE_STORAGE_ACCOUNT_KEY")
	v.BindEnv("azure.storage.connectionstring", "AZURE_STORAGE_CONNECTION_STRING")
	v.BindEnv("azure.storage.blobendpoint", "AZURE_STORAGE_BLOB_ENDPOINT")

	// Logging
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.format", "LOG_FORMAT")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate required fields
	if c.Database.URL == "" {
		return fmt.Errorf("database.url is required")
	}

	if c.Azure.OpenAI.Endpoint == "" {
		return fmt.Errorf("azure.openai.endpoint is required")
	}

	if c.Azure.OpenAI.APIKey == "" {
		return fmt.Errorf("azure.openai.apikey is required")
	}

	if c.Azure.OpenAI.Deployment == "" {
		return fmt.Errorf("azure.openai.deployment is required")
	}

	if c.Azure.Speech.SubscriptionKey == "" {
		return fmt.Errorf("azure.speech.subscriptionkey is required")
	}

	if c.Azure.Speech.Region == "" {
		return fmt.Errorf("azure.speech.region is required")
	}

	if c.Azure.Storage.ConnectionString == "" && (c.Azure.Storage.AccountName == "" || c.Azure.Storage.AccountKey == "") {
		return fmt.Errorf("azure storage credentials are required (either connection string or account name + key)")
	}

	return nil
}
