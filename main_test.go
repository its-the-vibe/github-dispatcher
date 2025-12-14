package main

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_CHANNEL")

	config := loadConfig()

	if config.RedisHost != "localhost" {
		t.Errorf("Expected RedisHost to be 'localhost', got '%s'", config.RedisHost)
	}

	if config.RedisPort != "6379" {
		t.Errorf("Expected RedisPort to be '6379', got '%s'", config.RedisPort)
	}

	if config.RedisChannel != "github-webhooks" {
		t.Errorf("Expected RedisChannel to be 'github-webhooks', got '%s'", config.RedisChannel)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	// Set environment variables
	os.Setenv("REDIS_HOST", "redis-server")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_CHANNEL", "custom-channel")

	config := loadConfig()

	if config.RedisHost != "redis-server" {
		t.Errorf("Expected RedisHost to be 'redis-server', got '%s'", config.RedisHost)
	}

	if config.RedisPort != "6380" {
		t.Errorf("Expected RedisPort to be '6380', got '%s'", config.RedisPort)
	}

	if config.RedisChannel != "custom-channel" {
		t.Errorf("Expected RedisChannel to be 'custom-channel', got '%s'", config.RedisChannel)
	}

	// Clean up
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_CHANNEL")
}

func TestGetEnv(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_VAR", "test_value")
	value := getEnv("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
	os.Unsetenv("TEST_VAR")

	// Test with environment variable not set
	value = getEnv("TEST_VAR", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}
}
