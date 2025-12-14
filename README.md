# github-dispatcher

A service which receives a github webhook notification from a redis pubsub channel and dispatches CI/CD operations.

## Overview

This is the first iteration of the GitHub Dispatcher service. Currently, it subscribes to a Redis pubsub channel and prints out GitHub webhook notification messages.

## Features

- Subscribe to Redis pubsub channels
- Receive and display GitHub webhook notifications
- Configurable via environment variables
- Docker Compose setup for easy deployment
- Graceful shutdown handling

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for containerized deployment)
- Redis server (provided via Docker Compose)

## Configuration

The service is configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis server hostname | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `REDIS_CHANNEL` | Redis pubsub channel to subscribe to | `github-webhooks` |

Copy `.env.example` to `.env` and adjust the values as needed:

```bash
cp .env.example .env
```

## Running Locally

### With Go

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the service:
   ```bash
   go run main.go
   ```

### With Docker Compose

1. Start all services:
   ```bash
   docker-compose up --build
   ```

2. Stop services:
   ```bash
   docker-compose down
   ```

## Testing

To test the service, publish a message to the Redis channel:

1. Connect to Redis CLI:
   ```bash
   # If using Docker Compose:
   docker exec -it github-dispatcher-redis redis-cli
   
   # If using local Redis:
   redis-cli
   ```

2. Publish a test message:
   ```bash
   PUBLISH github-webhooks "Test webhook message"
   ```

3. You should see the message printed in the dispatcher service logs.

## Development

### Building

```bash
go build -o github-dispatcher .
```

### Running Tests

```bash
go test ./...
```

## Architecture

The service consists of:
- **main.go**: Main service logic that connects to Redis and listens for pubsub messages
- **Docker Compose**: Orchestrates Redis and the dispatcher service
- **Dockerfile**: Multi-stage build for creating a minimal production image

## Future Enhancements

- Parse and validate GitHub webhook payloads
- Trigger CI/CD operations based on webhook events
- Add logging framework
- Add metrics and monitoring
- Implement webhook signature verification
