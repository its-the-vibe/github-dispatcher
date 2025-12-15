package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisHost         string
	RedisPort         string
	RedisChannel      string
	ConfigFilePath    string
	PipelineQueueName string
	LogLevel          string
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var currentLogLevel LogLevel = LogLevelInfo

type FilterRule struct {
	Repo     string   `json:"repo"`
	Branch   string   `json:"branch"`
	Type     string   `json:"type"`
	Dir      string   `json:"dir"`
	Commands []string `json:"commands"`
}

type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func loadConfig() Config {
	return Config{
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		RedisChannel:      getEnv("REDIS_CHANNEL", "github-webhook-push"),
		ConfigFilePath:    getEnv("CONFIG_FILE_PATH", "config.json"),
		PipelineQueueName: getEnv("PIPELINE_QUEUE_NAME", "pipeline"),
		LogLevel:          getEnv("LOG_LEVEL", "INFO"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

func logDebug(format string, v ...interface{}) {
	if currentLogLevel <= LogLevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func logInfo(format string, v ...interface{}) {
	if currentLogLevel <= LogLevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

func logWarn(format string, v ...interface{}) {
	if currentLogLevel <= LogLevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

func logError(format string, v ...interface{}) {
	if currentLogLevel <= LogLevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}

func loadFilterRules(filePath string) ([]FilterRule, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var rules []FilterRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return rules, nil
}

func findMatchingRule(rules []FilterRule, repo, branch string) *FilterRule {
	for i := range rules {
		if rules[i].Repo == repo && rules[i].Branch == branch {
			return &rules[i]
		}
	}
	return nil
}

func handleWebhookMessage(ctx context.Context, rdb *redis.Client, queueName string, rules []FilterRule, payload string) error {
	var event GitHubPushEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	logDebug("Processing push event for repo: %s, ref: %s", event.Repository.FullName, event.Ref)

	rule := findMatchingRule(rules, event.Repository.FullName, event.Ref)
	if rule == nil {
		logDebug("No matching rule found for repo: %s, ref: %s", event.Repository.FullName, event.Ref)
		return nil
	}

	logDebug("Found matching rule for repo: %s, ref: %s", rule.Repo, rule.Branch)

	// Serialize the matched rule to JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to serialize rule: %w", err)
	}

	// Push to Redis list
	if err := rdb.RPush(ctx, queueName, ruleJSON).Err(); err != nil {
		return fmt.Errorf("failed to push to Redis queue: %w", err)
	}

	logDebug("Pushed rule to queue '%s': %s", queueName, string(ruleJSON))
	return nil
}

func main() {
	config := loadConfig()
	currentLogLevel = parseLogLevel(config.LogLevel)

	logInfo("Starting GitHub Dispatcher Service...")
	logInfo("Configuration: Redis=%s:%s, Channel=%s, ConfigFile=%s, PipelineQueue=%s, LogLevel=%s",
		config.RedisHost, config.RedisPort, config.RedisChannel, config.ConfigFilePath, config.PipelineQueueName, config.LogLevel)

	// Load filter rules
	rules, err := loadFilterRules(config.ConfigFilePath)
	if err != nil {
		log.Fatalf("Failed to load filter rules: %v", err)
	}
	logInfo("Loaded %d filter rule(s)", len(rules))

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
	})
	defer rdb.Close()

	ctx := context.Background()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	logInfo("Successfully connected to Redis")

	// Subscribe to channel
	pubsub := rdb.Subscribe(ctx, config.RedisChannel)
	defer pubsub.Close()

	logInfo("Subscribed to channel: %s", config.RedisChannel)
	logInfo("Waiting for messages...")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel for receiving messages
	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			logDebug("Received message from channel '%s':\n%s", msg.Channel, msg.Payload)
			if err := handleWebhookMessage(ctx, rdb, config.PipelineQueueName, rules, msg.Payload); err != nil {
				logError("Error handling webhook message: %v", err)
			}
		case sig := <-sigChan:
			logInfo("Received signal: %v. Shutting down gracefully...", sig)
			return
		}
	}
}
