package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnvFileAndEnvOverride(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("API_PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("OBJECT_STORAGE_BUCKET", "")
	t.Setenv("TIKA_ENDPOINT", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := []byte("APP_ENV=staging\nAPI_PORT=9090\nLOG_LEVEL=debug\nDATABASE_DSN=postgres://from-file\nREDIS_ADDR=redis://file\nOBJECT_STORAGE_BUCKET=from-file\nTIKA_ENDPOINT=http://tika:9998\n")
	if err := os.WriteFile(envPath, content, 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("OBJECT_STORAGE_BUCKET", "from-env")

	cfg, err := Load(Options{EnvFile: envPath})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "staging" {
		t.Fatalf("AppEnv = %q, want staging", cfg.AppEnv)
	}
	if cfg.APIPort != "9090" {
		t.Fatalf("APIPort = %q, want 9090", cfg.APIPort)
	}
	if cfg.LogLevel != "warn" {
		t.Fatalf("LogLevel = %q, want warn", cfg.LogLevel)
	}
	if cfg.ObjectStorageBucket != "from-env" {
		t.Fatalf("ObjectStorageBucket = %q, want from-env", cfg.ObjectStorageBucket)
	}
}

func TestLoadDefaultsWhenOptionalValuesMissing(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("API_PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("OBJECT_STORAGE_BUCKET", "")
	t.Setenv("TIKA_ENDPOINT", "")

	cfg, err := Load(Options{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.APIPort != "8080" {
		t.Fatalf("APIPort = %q, want 8080", cfg.APIPort)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.DatabaseDSN != "" {
		t.Fatalf("DatabaseDSN = %q, want empty", cfg.DatabaseDSN)
	}
}

func TestSummaryMasksSensitiveValues(t *testing.T) {
	cfg := Config{
		AppEnv:              "development",
		APIPort:             "8080",
		LogLevel:            "debug",
		DatabaseDSN:         "postgres://user:secret@localhost:5432/app",
		RedisAddr:           "localhost:6379",
		ObjectStorageBucket: "bucket",
		TikaEndpoint:        "http://localhost:9998",
	}

	summary := cfg.Summary()

	if summary["database_dsn"] == cfg.DatabaseDSN {
		t.Fatalf("database dsn should be masked")
	}
	if summary["api_port"] != "8080" {
		t.Fatalf("api_port = %v, want 8080", summary["api_port"])
	}
}
