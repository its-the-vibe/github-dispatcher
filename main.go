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
}

type FilterRule struct {
	Repo     string   `json:"repo"`
	Branch   string   `json:"branch"`
	Type     string   `json:"type"`
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
		RedisChannel:      getEnv("REDIS_CHANNEL", "github-webhooks"),
		ConfigFilePath:    getEnv("CONFIG_FILE_PATH", "config.json"),
		PipelineQueueName: getEnv("PIPELINE_QUEUE_NAME", "pipeline"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
	for _, rule := range rules {
		if rule.Repo == repo && rule.Branch == branch {
			return &rule
		}
	}
	return nil
}

func handleWebhookMessage(ctx context.Context, rdb *redis.Client, queueName string, rules []FilterRule, payload string) error {
	var event GitHubPushEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	log.Printf("Processing push event for repo: %s, ref: %s", event.Repository.FullName, event.Ref)

	rule := findMatchingRule(rules, event.Repository.FullName, event.Ref)
	if rule == nil {
		log.Printf("No matching rule found for repo: %s, ref: %s", event.Repository.FullName, event.Ref)
		return nil
	}

	log.Printf("Found matching rule for repo: %s, ref: %s", rule.Repo, rule.Branch)

	// Serialize the matched rule to JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to serialize rule: %w", err)
	}

	// Push to Redis list
	if err := rdb.RPush(ctx, queueName, ruleJSON).Err(); err != nil {
		return fmt.Errorf("failed to push to Redis queue: %w", err)
	}

	log.Printf("Pushed rule to queue '%s': %s", queueName, string(ruleJSON))
	return nil
}

func main() {
	log.Println("Starting GitHub Dispatcher Service...")

	config := loadConfig()
	log.Printf("Configuration: Redis=%s:%s, Channel=%s, ConfigFile=%s, PipelineQueue=%s\n",
		config.RedisHost, config.RedisPort, config.RedisChannel, config.ConfigFilePath, config.PipelineQueueName)

	// Load filter rules
	rules, err := loadFilterRules(config.ConfigFilePath)
	if err != nil {
		log.Fatalf("Failed to load filter rules: %v", err)
	}
	log.Printf("Loaded %d filter rule(s)", len(rules))

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
	log.Println("Successfully connected to Redis")

	// Subscribe to channel
	pubsub := rdb.Subscribe(ctx, config.RedisChannel)
	defer pubsub.Close()

	log.Printf("Subscribed to channel: %s\n", config.RedisChannel)
	log.Println("Waiting for messages...")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel for receiving messages
	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			log.Printf("Received message from channel '%s':\n%s\n", msg.Channel, msg.Payload)
			if err := handleWebhookMessage(ctx, rdb, config.PipelineQueueName, rules, msg.Payload); err != nil {
				log.Printf("Error handling webhook message: %v", err)
			}
		case sig := <-sigChan:
			log.Printf("Received signal: %v. Shutting down gracefully...", sig)
			return
		}
	}
}
