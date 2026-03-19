package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validConfig() *Config {
	return &Config{
		AppPort:        "8080",
		DBHost:         "localhost",
		DBPort:         "5432",
		DBUser:         "postgres",
		DBName:         "testdb",
		JWTSecret:      "secret",
		GoogleClientID: "client-id",
	}
}

func TestValidate_AllFieldsSet(t *testing.T) {
	cfg := validConfig()
	err := cfg.validate()
	assert.NoError(t, err)
}

func TestValidate_MissingAppPort(t *testing.T) {
	cfg := validConfig()
	cfg.AppPort = ""
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "APP_PORT")
}

func TestValidate_MissingDBHost(t *testing.T) {
	cfg := validConfig()
	cfg.DBHost = ""
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB_HOST")
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWTSecret = ""
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestValidate_MissingGoogleClientID(t *testing.T) {
	cfg := validConfig()
	cfg.GoogleClientID = ""
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GOOGLE_CLIENT_ID")
}
