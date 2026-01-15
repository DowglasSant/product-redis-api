package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Keycloak KeycloakConfig
	App      AppConfig
}

type ServerConfig struct {
	Port            int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout     time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"10s"`
	WriteTimeout    time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
	ShutdownTimeout time.Duration `envconfig:"SERVER_SHUTDOWN_TIMEOUT" default:"30s"`
}

type DatabaseConfig struct {
	Host            string        `envconfig:"DB_HOST" default:"localhost"`
	Port            int           `envconfig:"DB_PORT" default:"5432"`
	User            string        `envconfig:"DB_USER" default:"postgres"`
	Password        string        `envconfig:"DB_PASSWORD" required:"true"`
	Name            string        `envconfig:"DB_NAME" default:"products_db"`
	SSLMode         string        `envconfig:"DB_SSLMODE" default:"disable"`
	MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

type RedisConfig struct {
	Host       string `envconfig:"REDIS_HOST" default:"localhost"`
	Port       int    `envconfig:"REDIS_PORT" default:"6379"`
	Password   string `envconfig:"REDIS_PASSWORD" required:"true"`
	DB         int    `envconfig:"REDIS_DB" default:"0"`
	MaxRetries int    `envconfig:"REDIS_MAX_RETRIES" default:"3"`
	PoolSize   int    `envconfig:"REDIS_POOL_SIZE" default:"10"`
}

type KeycloakConfig struct {
	URL      string `envconfig:"KEYCLOAK_URL" default:"http://localhost:8180"`
	Realm    string `envconfig:"KEYCLOAK_REALM" default:"product-api"`
	ClientID string `envconfig:"KEYCLOAK_CLIENT_ID" default:"product-api-client"`
}

type AppConfig struct {
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return &cfg, nil
}

func (c *DatabaseConfig) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production"
}

func (c *KeycloakConfig) JWKSURL() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.URL, c.Realm)
}

func (c *KeycloakConfig) Issuer() string {
	return fmt.Sprintf("%s/realms/%s", c.URL, c.Realm)
}
