package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"development"`
	Server ServerConfig
	DB     DBConfig
	Auth   AuthConfig
}

type ServerConfig struct {
	Port              string        `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout       time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"5s"`
	ReadHeaderTimeout time.Duration `env:"SERVER_READ_HEADER_TIMEOUT" envDefault:"2s"`
	WriteTimeout      time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout       time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"60s"`
}

type DBConfig struct {
	Host            string        `env:"DB_HOST,required"`
	Port            int           `env:"DB_PORT" envDefault:"5432"`
	User            string        `env:"DB_USER,required"`
	Password        string        `env:"DB_PASSWORD,required"`
	Name            string        `env:"DB_NAME,required"`
	SSLMode         string        `env:"DB_SSL_MODE" envDefault:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"10"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"10m"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"30m"`
	ConnTimeout     time.Duration `env:"DB_CONN_TIMEOUT" envDefault:"3s"`
}

type AuthConfig struct {
	JWTSecret      string        `env:"JWT_SECRET,required"`
	JWTExpiryHours time.Duration `env:"JWT_EXPIRY_HOURS" envDefault:"24h"`
}

func Load() *Config {
	_ = godotenv.Load()
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	return cfg
}
