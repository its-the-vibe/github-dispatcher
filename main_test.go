package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_CHANNEL")
	os.Unsetenv("CONFIG_FILE_PATH")
	os.Unsetenv("PIPELINE_QUEUE_NAME")
	os.Unsetenv("LOG_LEVEL")

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

	if config.ConfigFilePath != "config.json" {
		t.Errorf("Expected ConfigFilePath to be 'config.json', got '%s'", config.ConfigFilePath)
	}

	if config.PipelineQueueName != "pipeline" {
		t.Errorf("Expected PipelineQueueName to be 'pipeline', got '%s'", config.PipelineQueueName)
	}

	if config.LogLevel != "INFO" {
		t.Errorf("Expected LogLevel to be 'INFO', got '%s'", config.LogLevel)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	// Set environment variables
	os.Setenv("REDIS_HOST", "redis-server")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_CHANNEL", "custom-channel")
	os.Setenv("CONFIG_FILE_PATH", "/path/to/config.json")
	os.Setenv("PIPELINE_QUEUE_NAME", "custom-pipeline")
	os.Setenv("LOG_LEVEL", "DEBUG")

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

	if config.ConfigFilePath != "/path/to/config.json" {
		t.Errorf("Expected ConfigFilePath to be '/path/to/config.json', got '%s'", config.ConfigFilePath)
	}

	if config.PipelineQueueName != "custom-pipeline" {
		t.Errorf("Expected PipelineQueueName to be 'custom-pipeline', got '%s'", config.PipelineQueueName)
	}

	if config.LogLevel != "DEBUG" {
		t.Errorf("Expected LogLevel to be 'DEBUG', got '%s'", config.LogLevel)
	}

	// Clean up
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_CHANNEL")
	os.Unsetenv("CONFIG_FILE_PATH")
	os.Unsetenv("PIPELINE_QUEUE_NAME")
	os.Unsetenv("LOG_LEVEL")
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

func TestLoadFilterRules(t *testing.T) {
	// Create a temporary config file
	tempFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configData := `[
		{
			"repo": "owner/repo1",
			"branch": "refs/heads/main",
			"type": "git-webhook",
			"commands": ["make build", "make test"]
		},
		{
			"repo": "owner/repo2",
			"branch": "refs/heads/develop",
			"type": "git-webhook",
			"commands": ["npm install", "npm test"]
		}
	]`

	if _, err := tempFile.WriteString(configData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	rules, err := loadFilterRules(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load filter rules: %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}

	if rules[0].Repo != "owner/repo1" {
		t.Errorf("Expected repo 'owner/repo1', got '%s'", rules[0].Repo)
	}

	if rules[0].Branch != "refs/heads/main" {
		t.Errorf("Expected branch 'refs/heads/main', got '%s'", rules[0].Branch)
	}

	if len(rules[0].Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(rules[0].Commands))
	}
}

func TestLoadFilterRules_InvalidFile(t *testing.T) {
	_, err := loadFilterRules("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadFilterRules_InvalidJSON(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString("invalid json"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	_, err = loadFilterRules(tempFile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestFindMatchingRule(t *testing.T) {
	rules := []FilterRule{
		{
			Repo:     "owner/repo1",
			Branch:   "refs/heads/main",
			Type:     "git-webhook",
			Commands: []string{"make build"},
		},
		{
			Repo:     "owner/repo2",
			Branch:   "refs/heads/develop",
			Type:     "git-webhook",
			Commands: []string{"npm test"},
		},
	}

	// Test matching rule
	rule := findMatchingRule(rules, "owner/repo1", "refs/heads/main")
	if rule == nil {
		t.Error("Expected to find matching rule, got nil")
	} else {
		if rule.Repo != "owner/repo1" {
			t.Errorf("Expected repo 'owner/repo1', got '%s'", rule.Repo)
		}
		if rule.Branch != "refs/heads/main" {
			t.Errorf("Expected branch 'refs/heads/main', got '%s'", rule.Branch)
		}
	}

	// Test non-matching rule
	rule = findMatchingRule(rules, "owner/repo3", "refs/heads/main")
	if rule != nil {
		t.Error("Expected no matching rule, got one")
	}

	// Test with wrong branch
	rule = findMatchingRule(rules, "owner/repo1", "refs/heads/develop")
	if rule != nil {
		t.Error("Expected no matching rule for wrong branch, got one")
	}
}

func TestHandleWebhookMessage(t *testing.T) {
	rules := []FilterRule{
		{
			Repo:     "owner/test-repo",
			Branch:   "refs/heads/main",
			Type:     "git-webhook",
			Commands: []string{"make build"},
		},
	}

	// Test valid webhook payload
	payload := `{
		"ref": "refs/heads/main",
		"repository": {
			"full_name": "owner/test-repo"
		}
	}`

	var event GitHubPushEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("Failed to parse valid webhook payload: %v", err)
	}

	if event.Ref != "refs/heads/main" {
		t.Errorf("Expected ref 'refs/heads/main', got '%s'", event.Ref)
	}

	if event.Repository.FullName != "owner/test-repo" {
		t.Errorf("Expected repo 'owner/test-repo', got '%s'", event.Repository.FullName)
	}

	// Test finding the rule
	rule := findMatchingRule(rules, event.Repository.FullName, event.Ref)
	if rule == nil {
		t.Error("Expected to find matching rule, got nil")
	}

	// Test invalid payload
	invalidPayload := "not a json"
	var invalidEvent GitHubPushEvent
	if err := json.Unmarshal([]byte(invalidPayload), &invalidEvent); err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestHandleWebhookMessage_Integration(t *testing.T) {
	// Skip this test if Redis is not available
	// This is an integration test that requires a Redis instance
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Try to connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	// Use a unique queue name for this test to avoid conflicts
	queueName := "test-pipeline"
	
	// Clean up before test
	rdb.Del(ctx, queueName)
	defer rdb.Del(ctx, queueName)

	rules := []FilterRule{
		{
			Repo:     "owner/test-repo",
			Branch:   "refs/heads/main",
			Type:     "git-webhook",
			Commands: []string{"make build", "make test"},
		},
	}

	payload := `{
		"ref": "refs/heads/main",
		"repository": {
			"full_name": "owner/test-repo"
		}
	}`

	err := handleWebhookMessage(ctx, rdb, queueName, rules, payload)
	if err != nil {
		t.Fatalf("Failed to handle webhook message: %v", err)
	}

	// Verify the rule was pushed to Redis (FIFO: RPush adds to tail, LPop removes from head)
	result, err := rdb.LPop(ctx, queueName).Result()
	if err != nil {
		t.Fatalf("Failed to pop from Redis queue: %v", err)
	}

	var pushedRule FilterRule
	if err := json.Unmarshal([]byte(result), &pushedRule); err != nil {
		t.Fatalf("Failed to parse pushed rule: %v", err)
	}

	if pushedRule.Repo != "owner/test-repo" {
		t.Errorf("Expected repo 'owner/test-repo', got '%s'", pushedRule.Repo)
	}

	if pushedRule.Branch != "refs/heads/main" {
		t.Errorf("Expected branch 'refs/heads/main', got '%s'", pushedRule.Branch)
	}

	if len(pushedRule.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(pushedRule.Commands))
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", LogLevelDebug},
		{"INFO", LogLevelInfo},
		{"WARN", LogLevelWarn},
		{"WARNING", LogLevelWarn},
		{"ERROR", LogLevelError},
		{"invalid", LogLevelInfo}, // default to INFO
		{"", LogLevelInfo},         // default to INFO
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
