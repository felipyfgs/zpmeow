package config

import "time"

func DefaultConfig() *Config {
	return &Config{
		Database: DefaultDatabaseConfig(),
		Server:   DefaultServerConfig(),
		Auth:     DefaultAuthConfig(),
		Logging:  DefaultLoggingConfig(),
		CORS:     DefaultCORSConfig(),
		Webhook:  DefaultWebhookConfig(),
		Meow:     DefaultMeowConfig(),
		Security: DefaultSecurityConfig(),
	}
}

func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "",
		Name:            "meow",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:         "8080",
		Mode:         "debug",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		GlobalAPIKey:    "",
		SessionTimeout:  24 * time.Hour,
		TokenExpiration: 1 * time.Hour,
	}
}

func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:          "info",
		Format:         "console",
		ConsoleColor:   true,
		FileEnabled:    false,
		FilePath:       "log/app.log",
		FileMaxSize:    100,
		FileMaxBackups: 3,
		FileMaxAge:     28,
		FileCompress:   true,
		FileFormat:     "json",
	}
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowAllOrigins: true,
		AllowOrigins:    []string{},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key",
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

func DefaultWebhookConfig() WebhookConfig {
	return WebhookConfig{
		Timeout:           30 * time.Second,
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

func DefaultMeowConfig() MeowConfig {
	return MeowConfig{
		MaxRetries:        3,
		RetryInterval:     5 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		QRCodeTimeout:     60 * time.Second,
		ReconnectDelay:    10 * time.Second,
	}
}

func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		RateLimitEnabled: false,
		RateLimitRPS:     100,
		RequestTimeout:   30 * time.Second,
		MaxRequestSize:   10 * 1024 * 1024,
	}
}

func ProductionConfig() *Config {
	cfg := DefaultConfig()

	cfg.Server.Mode = "release"
	cfg.Server.ReadTimeout = 15 * time.Second
	cfg.Server.WriteTimeout = 15 * time.Second
	cfg.Server.IdleTimeout = 60 * time.Second

	cfg.Logging.Level = "warn"
	cfg.Logging.Format = "json"
	cfg.Logging.ConsoleColor = false
	cfg.Logging.FileEnabled = true
	cfg.Logging.FileFormat = "json"

	cfg.CORS.AllowAllOrigins = false
	cfg.CORS.AllowCredentials = true

	cfg.Security.RateLimitEnabled = true
	cfg.Security.RateLimitRPS = 50
	cfg.Security.RequestTimeout = 15 * time.Second
	cfg.Security.MaxRequestSize = 5 * 1024 * 1024

	cfg.Database.SSLMode = "require"
	cfg.Database.MaxOpenConns = 50
	cfg.Database.MaxIdleConns = 10
	cfg.Database.ConnMaxLifetime = 10 * time.Minute

	return cfg
}

func TestConfig() *Config {
	cfg := DefaultConfig()

	cfg.Server.Mode = "test"
	cfg.Server.Port = "0"

	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "console"
	cfg.Logging.FileEnabled = false

	cfg.Database.Name = "meow_test"
	cfg.Database.MaxOpenConns = 5
	cfg.Database.MaxIdleConns = 2

	cfg.Webhook.Timeout = 5 * time.Second
	cfg.Webhook.MaxRetries = 1

	cfg.Meow.ConnectionTimeout = 5 * time.Second
	cfg.Meow.QRCodeTimeout = 10 * time.Second
	cfg.Meow.MaxRetries = 1

	return cfg
}

func DevelopmentConfig() *Config {
	cfg := DefaultConfig()

	cfg.Server.Mode = "debug"

	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "console"
	cfg.Logging.ConsoleColor = true
	cfg.Logging.FileEnabled = false

	cfg.CORS.AllowAllOrigins = true
	cfg.CORS.AllowCredentials = false

	cfg.Security.RateLimitEnabled = false
	cfg.Security.RequestTimeout = 60 * time.Second
	cfg.Security.MaxRequestSize = 50 * 1024 * 1024

	return cfg
}
