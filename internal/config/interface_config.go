package config

import "time"

type ConfigProvider interface {
	GetDatabase() DatabaseConfigProvider
	GetServer() ServerConfigProvider
	GetAuth() AuthConfigProvider
	GetLogging() LoggingConfigProvider
	GetCORS() CORSConfigProvider
	GetWebhook() WebhookConfigProvider
	GetMeow() MeowConfigProvider
	GetSecurity() SecurityConfigProvider
	GetCache() CacheConfigProvider
}

type DatabaseConfigProvider interface {
	GetHost() string
	GetPort() string
	GetUser() string
	GetPassword() string
	GetName() string
	GetSSLMode() string
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() time.Duration
	GetURL() string
}

type ServerConfigProvider interface {
	GetPort() string
	GetMode() string
	GetReadTimeout() time.Duration
	GetWriteTimeout() time.Duration
	GetIdleTimeout() time.Duration
}

type AuthConfigProvider interface {
	GetGlobalAPIKey() string
	GetSessionTimeout() time.Duration
	GetTokenExpiration() time.Duration
}

type LoggingConfigProvider interface {
	GetLevel() string
	GetFormat() string
	GetConsoleColor() bool
	GetFileEnabled() bool
	GetFilePath() string
	GetFileMaxSize() int
	GetFileMaxBackups() int
	GetFileMaxAge() int
	GetFileCompress() bool
	GetFileFormat() string
}

type CORSConfigProvider interface {
	GetAllowAllOrigins() bool
	GetAllowOrigins() []string
	GetAllowMethods() []string
	GetAllowHeaders() []string
	GetExposeHeaders() []string
	GetAllowCredentials() bool
	GetMaxAge() int
}

type WebhookConfigProvider interface {
	GetTimeout() time.Duration
	GetMaxRetries() int
	GetInitialBackoff() time.Duration
	GetMaxBackoff() time.Duration
	GetBackoffMultiplier() float64
}

type MeowConfigProvider interface {
	GetMaxRetries() int
	GetRetryInterval() time.Duration
	GetConnectionTimeout() time.Duration
	GetQRCodeTimeout() time.Duration
	GetReconnectDelay() time.Duration
}

type SecurityConfigProvider interface {
	GetRateLimitEnabled() bool
	GetRateLimitRPS() int
	GetRequestTimeout() time.Duration
	GetMaxRequestSize() int64
}

type CacheConfigProvider interface {
	GetCacheEnabled() bool
	GetRedisURL() string
	GetRedisHost() string
	GetRedisPort() string
	GetRedisPassword() string
	GetRedisDB() int
	GetPoolSize() int
	GetMinIdleConns() int
	GetMaxRetries() int
	GetRetryDelay() time.Duration
	GetDialTimeout() time.Duration
	GetReadTimeout() time.Duration
	GetWriteTimeout() time.Duration
	GetSessionTTL() time.Duration
	GetQRCodeTTL() time.Duration
	GetCredentialTTL() time.Duration
	GetStatusTTL() time.Duration
}

func (c *Config) GetDatabase() DatabaseConfigProvider {
	return &c.Database
}

func (c *Config) GetServer() ServerConfigProvider {
	return &c.Server
}

func (c *Config) GetAuth() AuthConfigProvider {
	return &c.Auth
}

func (c *Config) GetLogging() LoggingConfigProvider {
	return &c.Logging
}

func (c *Config) GetCORS() CORSConfigProvider {
	return &c.CORS
}

func (c *Config) GetWebhook() WebhookConfigProvider {
	return &c.Webhook
}

func (c *Config) GetMeow() MeowConfigProvider {
	return &c.Meow
}

func (c *Config) GetSecurity() SecurityConfigProvider {
	return &c.Security
}

func (c *Config) GetCache() CacheConfigProvider {
	return &c.Cache
}

func (d *DatabaseConfig) GetHost() string                   { return d.Host }
func (d *DatabaseConfig) GetPort() string                   { return d.Port }
func (d *DatabaseConfig) GetUser() string                   { return d.User }
func (d *DatabaseConfig) GetPassword() string               { return d.Password }
func (d *DatabaseConfig) GetName() string                   { return d.Name }
func (d *DatabaseConfig) GetSSLMode() string                { return d.SSLMode }
func (d *DatabaseConfig) GetMaxOpenConns() int              { return d.MaxOpenConns }
func (d *DatabaseConfig) GetMaxIdleConns() int              { return d.MaxIdleConns }
func (d *DatabaseConfig) GetConnMaxLifetime() time.Duration { return d.ConnMaxLifetime }
func (d *DatabaseConfig) GetURL() string                    { return d.URL }

func (s *ServerConfig) GetPort() string                { return s.Port }
func (s *ServerConfig) GetMode() string                { return s.Mode }
func (s *ServerConfig) GetReadTimeout() time.Duration  { return s.ReadTimeout }
func (s *ServerConfig) GetWriteTimeout() time.Duration { return s.WriteTimeout }
func (s *ServerConfig) GetIdleTimeout() time.Duration  { return s.IdleTimeout }

func (a *AuthConfig) GetGlobalAPIKey() string           { return a.GlobalAPIKey }
func (a *AuthConfig) GetSessionTimeout() time.Duration  { return a.SessionTimeout }
func (a *AuthConfig) GetTokenExpiration() time.Duration { return a.TokenExpiration }

func (l *LoggingConfig) GetLevel() string       { return l.Level }
func (l *LoggingConfig) GetFormat() string      { return l.Format }
func (l *LoggingConfig) GetConsoleColor() bool  { return l.ConsoleColor }
func (l *LoggingConfig) GetFileEnabled() bool   { return l.FileEnabled }
func (l *LoggingConfig) GetFilePath() string    { return l.FilePath }
func (l *LoggingConfig) GetFileMaxSize() int    { return l.FileMaxSize }
func (l *LoggingConfig) GetFileMaxBackups() int { return l.FileMaxBackups }
func (l *LoggingConfig) GetFileMaxAge() int     { return l.FileMaxAge }
func (l *LoggingConfig) GetFileCompress() bool  { return l.FileCompress }
func (l *LoggingConfig) GetFileFormat() string  { return l.FileFormat }

func (c *CORSConfig) GetAllowAllOrigins() bool   { return c.AllowAllOrigins }
func (c *CORSConfig) GetAllowOrigins() []string  { return c.AllowOrigins }
func (c *CORSConfig) GetAllowMethods() []string  { return c.AllowMethods }
func (c *CORSConfig) GetAllowHeaders() []string  { return c.AllowHeaders }
func (c *CORSConfig) GetExposeHeaders() []string { return c.ExposeHeaders }
func (c *CORSConfig) GetAllowCredentials() bool  { return c.AllowCredentials }
func (c *CORSConfig) GetMaxAge() int             { return c.MaxAge }

func (w *WebhookConfig) GetTimeout() time.Duration        { return w.Timeout }
func (w *WebhookConfig) GetMaxRetries() int               { return w.MaxRetries }
func (w *WebhookConfig) GetInitialBackoff() time.Duration { return w.InitialBackoff }
func (w *WebhookConfig) GetMaxBackoff() time.Duration     { return w.MaxBackoff }
func (w *WebhookConfig) GetBackoffMultiplier() float64    { return w.BackoffMultiplier }

func (w *MeowConfig) GetMaxRetries() int                  { return w.MaxRetries }
func (w *MeowConfig) GetRetryInterval() time.Duration     { return w.RetryInterval }
func (w *MeowConfig) GetConnectionTimeout() time.Duration { return w.ConnectionTimeout }
func (w *MeowConfig) GetQRCodeTimeout() time.Duration     { return w.QRCodeTimeout }
func (w *MeowConfig) GetReconnectDelay() time.Duration    { return w.ReconnectDelay }

func (s *SecurityConfig) GetRateLimitEnabled() bool        { return s.RateLimitEnabled }
func (s *SecurityConfig) GetRateLimitRPS() int             { return s.RateLimitRPS }
func (s *SecurityConfig) GetRequestTimeout() time.Duration { return s.RequestTimeout }
func (s *SecurityConfig) GetMaxRequestSize() int64         { return s.MaxRequestSize }

func (c *CacheConfig) GetCacheEnabled() bool           { return c.Enabled }
func (c *CacheConfig) GetRedisURL() string             { return c.RedisURL }
func (c *CacheConfig) GetRedisHost() string            { return c.RedisHost }
func (c *CacheConfig) GetRedisPort() string            { return c.RedisPort }
func (c *CacheConfig) GetRedisPassword() string        { return c.RedisPassword }
func (c *CacheConfig) GetRedisDB() int                 { return c.RedisDB }
func (c *CacheConfig) GetPoolSize() int                { return c.PoolSize }
func (c *CacheConfig) GetMinIdleConns() int            { return c.MinIdleConns }
func (c *CacheConfig) GetMaxRetries() int              { return c.MaxRetries }
func (c *CacheConfig) GetRetryDelay() time.Duration    { return c.RetryDelay }
func (c *CacheConfig) GetDialTimeout() time.Duration   { return c.DialTimeout }
func (c *CacheConfig) GetReadTimeout() time.Duration   { return c.ReadTimeout }
func (c *CacheConfig) GetWriteTimeout() time.Duration  { return c.WriteTimeout }
func (c *CacheConfig) GetSessionTTL() time.Duration    { return c.SessionTTL }
func (c *CacheConfig) GetQRCodeTTL() time.Duration     { return c.QRCodeTTL }
func (c *CacheConfig) GetCredentialTTL() time.Duration { return c.CredentialTTL }
func (c *CacheConfig) GetStatusTTL() time.Duration     { return c.StatusTTL }
