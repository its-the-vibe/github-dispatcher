package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisHost    string
	RedisPort    string
	RedisChannel string
}

func loadConfig() Config {
	return Config{
		RedisHost:    getEnv("REDIS_HOST", "localhost"),
		RedisPort:    getEnv("REDIS_PORT", "6379"),
		RedisChannel: getEnv("REDIS_CHANNEL", "github-webhooks"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	log.Println("Starting GitHub Dispatcher Service...")

	config := loadConfig()
	log.Printf("Configuration: Redis=%s:%s, Channel=%s\n", 
		config.RedisHost, config.RedisPort, config.RedisChannel)

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
		case sig := <-sigChan:
			log.Printf("Received signal: %v. Shutting down gracefully...", sig)
			return
		}
	}
}
