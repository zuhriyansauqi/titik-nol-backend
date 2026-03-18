package config

import (
	"log"
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
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
