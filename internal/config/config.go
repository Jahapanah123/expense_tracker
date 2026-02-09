package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	databaseURL           string
	dBMaxOpenConns        int
	dBMaxIdleConns        int
	dbIdleConnLifetime    time.Duration
	dBConnMaxLifeTime     time.Duration
	dBConnTimeOut         time.Duration
	port                  string
	serverReadTimeOut     time.Duration
	serverWriteTimeOut    time.Duration
	serverIdleTimeOut     time.Duration
	serverShutDownTimeOut time.Duration
	jwtSecret             string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // silently ignore if .env not found

	cfg := &Config{
		databaseURL:           getEnv("DATABASE_URL", ""),
		dBMaxOpenConns:        getEnvInt("DB_MAX_OPEN_CONNS", 25),
		dBMaxIdleConns:        getEnvInt("DB_MAX_IDLE_CONNS", 10),
		dbIdleConnLifetime:    getEnvTimeDuration("DB_IDLE_CONN_LIFETIME", 10*time.Second),
		dBConnMaxLifeTime:     getEnvTimeDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		dBConnTimeOut:         getEnvTimeDuration("DB_CONN_TIMEOUT", 5*time.Second),
		serverReadTimeOut:     getEnvTimeDuration("SERVER_READ_TIMEOUT", 5*time.Second),
		serverWriteTimeOut:    getEnvTimeDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		serverIdleTimeOut:     getEnvTimeDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		serverShutDownTimeOut: getEnvTimeDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second),
		port:                  getEnv("PORT", ""),
		jwtSecret:             getEnv("JWT_SECRET", ""),
	}

	// Validate required fields
	if cfg.databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Default port
	if cfg.port == "" {
		cfg.port = "8080"
	}

	return cfg, nil
}

// Get functions to read values
func (c *Config) DatabaseURL() string                  { return c.databaseURL }
func (c *Config) DBMaxOpenConns() int                  { return c.dBMaxOpenConns }
func (c *Config) DBMaxIdleConns() int                  { return c.dBMaxIdleConns }
func (c *Config) DBIdleConnLifetime() time.Duration    { return c.dbIdleConnLifetime }
func (c *Config) DBConnMaxLifeTime() time.Duration     { return c.dBConnMaxLifeTime }
func (c *Config) DBConnTimeOut() time.Duration         { return c.dBConnTimeOut }
func (c *Config) Port() string                         { return c.port }
func (c *Config) ServerReadTimeOut() time.Duration     { return c.serverReadTimeOut }
func (c *Config) ServerWriteTimeOut() time.Duration    { return c.serverWriteTimeOut }
func (c *Config) ServerIdleTimeOut() time.Duration     { return c.serverIdleTimeOut }
func (c *Config) ServerShutDownTimeOut() time.Duration { return c.serverShutDownTimeOut }
func (c *Config) JwtSecret() string                    { return c.jwtSecret }

// function to get string from Getenv
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// function get get int from Getenv
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil { // every value comes from env is string so change to int
			return v
		}
	}
	return defaultValue
}

// function to get time.duration from Getenv

func getEnvTimeDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if v, err := time.ParseDuration(value); err == nil {
			return v
		}
	}
	return defaultValue
}
