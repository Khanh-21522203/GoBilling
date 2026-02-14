package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env  string `envconfig:"GOBILLING_ENV" default:"development"`
	Port int    `envconfig:"PORT" default:"8080"`

	Database DatabaseConfig
	Redis    RedisConfig
	Secrets  SecretsConfig
	Logging  LoggingConfig
	Otel     OtelConfig
	Metrics  MetricsConfig
	RateLimit RateLimitConfig
	Payment  PaymentConfig
}

type DatabaseConfig struct {
	Host        string `envconfig:"DB_HOST" default:"localhost"`
	Port        int    `envconfig:"DB_PORT" default:"5432"`
	Name        string `envconfig:"DB_NAME" default:"gobilling"`
	User        string `envconfig:"DB_USER" default:"gobilling_app"`
	Password    string `envconfig:"DB_PASSWORD" default:"changeme"`
	SSLMode     string `envconfig:"DB_SSL_MODE" default:"disable"`
	MaxConns    int32  `envconfig:"DB_MAX_CONNS" default:"25"`
	MinConns    int32  `envconfig:"DB_MIN_CONNS" default:"5"`
	MaxConnLife time.Duration `envconfig:"DB_MAX_CONN_LIFETIME" default:"1h"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	URL      string `envconfig:"REDIS_URL" default:"redis://localhost:6379/0"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
}

type SecretsConfig struct {
	Backend string `envconfig:"SECRETS_BACKEND" default:"env"`
}

type LoggingConfig struct {
	Level  string `envconfig:"LOG_LEVEL" default:"info"`
	Format string `envconfig:"LOG_FORMAT" default:"json"`
}

type OtelConfig struct {
	Enabled        bool   `envconfig:"OTEL_ENABLED" default:"false"`
	JaegerEndpoint string `envconfig:"OTEL_EXPORTER_JAEGER_ENDPOINT" default:"http://localhost:14268/api/traces"`
}

type MetricsConfig struct {
	Enabled bool `envconfig:"METRICS_ENABLED" default:"true"`
}

type RateLimitConfig struct {
	ReadPerMinute  int `envconfig:"RATE_LIMIT_READ_PER_MINUTE" default:"1000"`
	WritePerMinute int `envconfig:"RATE_LIMIT_WRITE_PER_MINUTE" default:"200"`
}

type PaymentConfig struct {
	Provider    string  `envconfig:"PAYMENT_PROVIDER" default:"stub"`
	SuccessRate float64 `envconfig:"PAYMENT_PROVIDER_SUCCESS_RATE" default:"0.9"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
