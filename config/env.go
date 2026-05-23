package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var once sync.Once

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		os.Setenv(key, value)
	}
}

func ensureEnvLoaded() {
	once.Do(loadEnvFile)
}

func GetString(key, fallback string) string {
	ensureEnvLoaded()
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	ensureEnvLoaded()
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func GetDuration(key string, fallback time.Duration) time.Duration {
	ensureEnvLoaded()
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func GetBool(key string, fallback bool) bool {
	ensureEnvLoaded()
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "true" || value == "1" || value == "yes" {
		return true
	}
	if value == "false" || value == "0" || value == "no" {
		return false
	}
	return fallback
}
