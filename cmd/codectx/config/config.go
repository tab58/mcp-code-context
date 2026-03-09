package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Default configuration values.
const (
	DefaultFalkorDBPort  = "6379"
	DefaultFalkorDBGraph = "codecontext"
	DefaultMCPPort       = "8080"
)

type Config struct {
	FalkorDBHost     string `env:"FALKORDB_HOST"`
	FalkorDBPort     string `env:"FALKORDB_PORT"`
	FalkorDBPassword string `env:"FALKORDB_PASSWORD"`
	FalkorDBGraph    string `env:"FALKORDB_GRAPH"`
	MCPPort          string `env:"MCP_PORT"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Ignore missing .env file — env vars may be set directly
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	host, err := requireEnv("FALKORDB_HOST")
	if err != nil {
		return nil, err
	}

	password, err := requireEnv("FALKORDB_PASSWORD")
	if err != nil {
		return nil, err
	}

	return &Config{
		FalkorDBHost:     host,
		FalkorDBPort:     envOrDefault("FALKORDB_PORT", DefaultFalkorDBPort),
		FalkorDBPassword: password,
		FalkorDBGraph:    envOrDefault("FALKORDB_GRAPH", DefaultFalkorDBGraph),
		MCPPort:          envOrDefault("MCP_PORT", DefaultMCPPort),
	}, nil
}

// requireEnv reads an environment variable and returns an error if it is empty.
func requireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return val, nil
}

// envOrDefault reads an environment variable and returns fallback if empty.
func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
