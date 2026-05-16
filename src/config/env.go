package config

import (
	"fmt"
	"os"
)

func GetEnv(key string) string {
	return os.Getenv(key)
}

func RequireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required env variable %q is not set", key)
	}
	return v, nil
}

func GetEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
