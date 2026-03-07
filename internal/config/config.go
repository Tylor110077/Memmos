package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Options struct {
	EnvFile string
}

type Config struct {
	AppEnv              string
	APIPort             string
	LogLevel            string
	DatabaseDSN         string
	RedisAddr           string
	ObjectStorageBucket string
	TikaEndpoint        string
}

func Load(opts Options) (Config, error) {
	values := map[string]string{}

	if opts.EnvFile != "" {
		fileValues, err := loadEnvFile(opts.EnvFile)
		if err != nil {
			return Config{}, err
		}
		for key, value := range fileValues {
			values[key] = value
		}
	}

	for _, key := range []string{
		"APP_ENV",
		"API_PORT",
		"LOG_LEVEL",
		"DATABASE_DSN",
		"REDIS_ADDR",
		"OBJECT_STORAGE_BUCKET",
		"TIKA_ENDPOINT",
	} {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			values[key] = value
		}
	}

	return Config{
		AppEnv:              fallback(values["APP_ENV"], "development"),
		APIPort:             fallback(values["API_PORT"], "8080"),
		LogLevel:            fallback(values["LOG_LEVEL"], "info"),
		DatabaseDSN:         values["DATABASE_DSN"],
		RedisAddr:           values["REDIS_ADDR"],
		ObjectStorageBucket: values["OBJECT_STORAGE_BUCKET"],
		TikaEndpoint:        values["TIKA_ENDPOINT"],
	}, nil
}

func (c Config) Summary() map[string]any {
	return map[string]any{
		"app_env":               c.AppEnv,
		"api_port":              c.APIPort,
		"log_level":             c.LogLevel,
		"database_dsn":          mask(c.DatabaseDSN),
		"redis_addr":            c.RedisAddr,
		"object_storage_bucket": c.ObjectStorageBucket,
		"tika_endpoint":         c.TikaEndpoint,
	}
}

func loadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("open env file: %w", err)
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid env line: %q", line)
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file: %w", err)
	}
	return values, nil
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func mask(value string) string {
	if value == "" {
		return ""
	}
	return "[redacted]"
}
