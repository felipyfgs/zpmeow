package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	Auth     AuthConfig     `json:"auth"`
	Logging  LoggingConfig  `json:"logging"`
	CORS     CORSConfig     `json:"cors"`
	Webhook  WebhookConfig  `json:"webhook"`
	Meow     MeowConfig     `json:"meow"`
	Security SecurityConfig `json:"security"`
}

type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            string        `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Name            string        `json:"name"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	URL             string        `json:"url"` // Computed field
}

type ServerConfig struct {
	Port         string        `json:"port"`
	Mode         string        `json:"mode"` // debug, release, test
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type AuthConfig struct {
	GlobalAPIKey    string        `json:"global_api_key"`
	SessionTimeout  time.Duration `json:"session_timeout"`
	TokenExpiration time.Duration `json:"token_expiration"`
}

type LoggingConfig struct {
	Level          string `json:"level"`
	Format         string `json:"format"`
	ConsoleColor   bool   `json:"console_color"`
	FileEnabled    bool   `json:"file_enabled"`
	FilePath       string `json:"file_path"`
	FileMaxSize    int    `json:"file_max_size"`
	FileMaxBackups int    `json:"file_max_backups"`
	FileMaxAge     int    `json:"file_max_age"`
	FileCompress   bool   `json:"file_compress"`
	FileFormat     string `json:"file_format"`
}

type CORSConfig struct {
	AllowAllOrigins  bool     `json:"allow_all_origins"`
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

type WebhookConfig struct {
	Timeout           time.Duration `json:"timeout"`
	MaxRetries        int           `json:"max_retries"`
	InitialBackoff    time.Duration `json:"initial_backoff"`
	MaxBackoff        time.Duration `json:"max_backoff"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

type MeowConfig struct {
	MaxRetries        int           `json:"max_retries"`
	RetryInterval     time.Duration `json:"retry_interval"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	QRCodeTimeout     time.Duration `json:"qr_code_timeout"`
	ReconnectDelay    time.Duration `json:"reconnect_delay"`
}

type SecurityConfig struct {
	RateLimitEnabled bool          `json:"rate_limit_enabled"`
	RateLimitRPS     int           `json:"rate_limit_rps"`
	RequestTimeout   time.Duration `json:"request_timeout"`
	MaxRequestSize   int64         `json:"max_request_size"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Database: loadDatabaseConfig(),
		Server:   loadServerConfig(),
		Auth:     loadAuthConfig(),
		Logging:  loadLoggingConfig(),
		CORS:     loadCORSConfig(),
		Webhook:  loadWebhookConfig(),
		Meow:     loadMeowConfig(),
		Security: loadSecurityConfig(),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Auth.GlobalAPIKey == "" {
		return fmt.Errorf("global API key is required")
	}
	return nil
}

func (c *Config) GetDatabaseURL() string {
	return c.Database.URL
}

func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "debug"
}

func (c *Config) IsTest() bool {
	return c.Server.Mode == "test"
}

func loadDatabaseConfig() DatabaseConfig {
	cfg := DatabaseConfig{
		Host:            getEnvOrDefault("DB_HOST", "localhost"),
		Port:            getEnvOrDefault("DB_PORT", "5432"),
		User:            getEnvOrDefault("DB_USER", "postgres"),
		Password:        getEnvOrDefault("DB_PASSWORD", ""),
		Name:            getEnvOrDefault("DB_NAME", "meow"),
		SSLMode:         getEnvOrDefault("DB_SSLMODE", "disable"),
		MaxOpenConns:    getIntEnvOrDefault("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getIntEnvOrDefault("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getDurationEnvOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}

	cfg.URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	return cfg
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port:         getEnvOrDefault("SERVER_PORT", "8080"),
		Mode:         getEnvOrDefault("GIN_MODE", "debug"),
		ReadTimeout:  getDurationEnvOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getDurationEnvOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getDurationEnvOrDefault("SERVER_IDLE_TIMEOUT", 120*time.Second),
	}
}

func loadAuthConfig() AuthConfig {
	return AuthConfig{
		GlobalAPIKey:    getEnvOrDefault("GLOBAL_API_KEY", ""),
		SessionTimeout:  getDurationEnvOrDefault("AUTH_SESSION_TIMEOUT", 24*time.Hour),
		TokenExpiration: getDurationEnvOrDefault("AUTH_TOKEN_EXPIRATION", 1*time.Hour),
	}
}

func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:          getEnvOrDefault("LOG_LEVEL", "info"),
		Format:         getEnvOrDefault("LOG_FORMAT", "console"),
		ConsoleColor:   getBoolEnvOrDefault("LOG_CONSOLE_COLOR", true),
		FileEnabled:    getBoolEnvOrDefault("LOG_FILE_ENABLED", true),
		FilePath:       getEnvOrDefault("LOG_FILE_PATH", "log/app.log"),
		FileMaxSize:    getIntEnvOrDefault("LOG_FILE_MAX_SIZE", 100),
		FileMaxBackups: getIntEnvOrDefault("LOG_FILE_MAX_BACKUPS", 3),
		FileMaxAge:     getIntEnvOrDefault("LOG_FILE_MAX_AGE", 28),
		FileCompress:   getBoolEnvOrDefault("LOG_FILE_COMPRESS", true),
		FileFormat:     getEnvOrDefault("LOG_FILE_FORMAT", "json"),
	}
}

func loadCORSConfig() CORSConfig {
	return CORSConfig{
		AllowAllOrigins:  getBoolEnvOrDefault("CORS_ALLOW_ALL_ORIGINS", true),
		AllowOrigins:     getStringSliceEnvOrDefault("CORS_ALLOW_ORIGINS", []string{}),
		AllowMethods:     getStringSliceEnvOrDefault("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}),
		AllowHeaders:     getStringSliceEnvOrDefault("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"}),
		ExposeHeaders:    getStringSliceEnvOrDefault("CORS_EXPOSE_HEADERS", []string{}),
		AllowCredentials: getBoolEnvOrDefault("CORS_ALLOW_CREDENTIALS", false),
		MaxAge:           getIntEnvOrDefault("CORS_MAX_AGE", 86400),
	}
}

func loadWebhookConfig() WebhookConfig {
	return WebhookConfig{
		Timeout:           getDurationEnvOrDefault("WEBHOOK_TIMEOUT", 30*time.Second),
		MaxRetries:        getIntEnvOrDefault("WEBHOOK_MAX_RETRIES", 3),
		InitialBackoff:    getDurationEnvOrDefault("WEBHOOK_INITIAL_BACKOFF", 1*time.Second),
		MaxBackoff:        getDurationEnvOrDefault("WEBHOOK_MAX_BACKOFF", 30*time.Second),
		BackoffMultiplier: getFloat64EnvOrDefault("WEBHOOK_BACKOFF_MULTIPLIER", 2.0),
	}
}

func loadMeowConfig() MeowConfig {
	return MeowConfig{
		MaxRetries:        getIntEnvOrDefault("meow_MAX_RETRIES", 3),
		RetryInterval:     getDurationEnvOrDefault("meow_RETRY_INTERVAL", 5*time.Second),
		ConnectionTimeout: getDurationEnvOrDefault("meow_CONNECTION_TIMEOUT", 30*time.Second),
		QRCodeTimeout:     getDurationEnvOrDefault("meow_QR_CODE_TIMEOUT", 60*time.Second),
		ReconnectDelay:    getDurationEnvOrDefault("meow_RECONNECT_DELAY", 10*time.Second),
	}
}

func loadSecurityConfig() SecurityConfig {
	return SecurityConfig{
		RateLimitEnabled: getBoolEnvOrDefault("SECURITY_RATE_LIMIT_ENABLED", false),
		RateLimitRPS:     getIntEnvOrDefault("SECURITY_RATE_LIMIT_RPS", 100),
		RequestTimeout:   getDurationEnvOrDefault("SECURITY_REQUEST_TIMEOUT", 30*time.Second),
		MaxRequestSize:   getInt64EnvOrDefault("SECURITY_MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getInt64EnvOrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getFloat64EnvOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getStringSliceEnvOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
