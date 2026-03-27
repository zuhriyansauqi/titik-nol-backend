package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string `mapstructure:"APP_NAME"`
	AppEnv  string `mapstructure:"APP_ENV"`
	AppPort string `mapstructure:"APP_PORT"`

	DBHost            string        `mapstructure:"DB_HOST"`
	DBPort            string        `mapstructure:"DB_PORT"`
	DBUser            string        `mapstructure:"DB_USER"`
	DBPassword        string        `mapstructure:"DB_PASSWORD"`
	DBName            string        `mapstructure:"DB_NAME"`
	DBSSLMode         string        `mapstructure:"DB_SSLMODE"`
	DBTimezone        string        `mapstructure:"DB_TIMEZONE"`
	DBMaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`

	LogLevel  string `mapstructure:"LOG_LEVEL"`
	LogFormat string `mapstructure:"LOG_FORMAT"`

	JWTSecret        string `mapstructure:"JWT_SECRET"`
	JWTIssuer        string `mapstructure:"JWT_ISSUER"`
	JWTExpirySeconds int    `mapstructure:"JWT_EXPIRY_SECONDS"`
	GoogleClientID   string `mapstructure:"GOOGLE_CLIENT_ID"`

	CORSAllowOrigins string  `mapstructure:"CORS_ALLOW_ORIGINS"`
	RateLimitRPS     float64 `mapstructure:"RATE_LIMIT_RPS"`
	RateLimitBurst   int     `mapstructure:"RATE_LIMIT_BURST"`

	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	CacheEnabled      bool          `mapstructure:"CACHE_ENABLED"`
	CacheDefaultTTL   time.Duration `mapstructure:"CACHE_DEFAULT_TTL"`
	CacheCategoryTTL  time.Duration `mapstructure:"CACHE_CATEGORY_TTL"`
	CacheDashboardTTL time.Duration `mapstructure:"CACHE_DASHBOARD_TTL"`
	CacheUserTTL      time.Duration `mapstructure:"CACHE_USER_TTL"`
	CacheAccountTTL   time.Duration `mapstructure:"CACHE_ACCOUNT_TTL"`
}

func (c *Config) validate() error {
	required := map[string]string{
		"APP_PORT":         c.AppPort,
		"DB_HOST":          c.DBHost,
		"DB_PORT":          c.DBPort,
		"DB_USER":          c.DBUser,
		"DB_NAME":          c.DBName,
		"JWT_SECRET":       c.JWTSecret,
		"GOOGLE_CLIENT_ID": c.GoogleClientID,
	}

	for key, val := range required {
		if val == "" {
			return fmt.Errorf("required config %s is not set", key)
		}
	}

	return nil
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Redis defaults
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)

	// Cache defaults
	viper.SetDefault("CACHE_ENABLED", true)
	viper.SetDefault("CACHE_DEFAULT_TTL", 5*time.Minute)
	viper.SetDefault("CACHE_CATEGORY_TTL", 30*time.Minute)
	viper.SetDefault("CACHE_DASHBOARD_TTL", 2*time.Minute)
	viper.SetDefault("CACHE_USER_TTL", 10*time.Minute)
	viper.SetDefault("CACHE_ACCOUNT_TTL", 5*time.Minute)

	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("Warning: .env file not found, using environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

