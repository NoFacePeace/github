package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port         int
	Environment  string
	ReadTimeout  int
	WriteTimeout int
}

func Load() (*Config, error) {
	port, err := getEnvInt("SERVER_PORT", 8080)
	if err != nil {
		return nil, err
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	readTimeout, err := getEnvInt("READ_TIMEOUT", 10)
	if err != nil {
		return nil, err
	}

	writeTimeout, err := getEnvInt("WRITE_TIMEOUT", 10)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:         port,
		Environment:  env,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}, nil
}

func getEnvInt(key string, fallback int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}
