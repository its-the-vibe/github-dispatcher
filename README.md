# github-dispatcher

A service which receives a github webhook notification from a redis pubsub channel and dispatches CI/CD operations.

## Overview

This GitHub Dispatcher service subscribes to a Redis pubsub channel, receives GitHub webhook notifications, filters them based on repository and branch configurations, and pushes matching configurations to a Redis queue for processing by CI/CD pipelines.

## Features

- Subscribe to Redis pubsub channels
- Receive and parse GitHub push webhook notifications
- Filter webhooks by repository name and branch
- Push matched configurations to Redis queue for pipeline processing
- Configurable via environment variables and JSON configuration file
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
| `CONFIG_FILE_PATH` | Path to the filter configuration JSON file | `config.json` |
| `PIPELINE_QUEUE_NAME` | Redis queue name for pushing matched configurations | `pipeline` |

Copy `.env.example` to `.env` and adjust the values as needed:

```bash
cp .env.example .env
```

### Filter Configuration File

Create a `config.json` file to define which repositories and branches should trigger CI/CD operations:

```json
[
  {
    "repo": "owner/repository-name",
    "branch": "refs/heads/main",
    "type": "git-webhook",
    "commands": ["make build", "make test", "make deploy"]
  },
  {
    "repo": "owner/another-repo",
    "branch": "refs/heads/develop",
    "type": "git-webhook",
    "commands": ["npm install", "npm run build", "npm test"]
  }
]
```

See `config.json.example` for a sample configuration file.

**Configuration Fields:**
- `repo`: Full repository name (e.g., `owner/repository-name`)
- `branch`: Branch reference to match (e.g., `refs/heads/main`)
- `type`: Type of webhook (currently `git-webhook`)
- `commands`: Array of CI/CD commands to execute

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

### Testing with a GitHub Webhook

To test the service with a GitHub push webhook, publish a message to the Redis channel:

1. Connect to Redis CLI:
   ```bash
   # If using Docker Compose:
   docker exec -it github-dispatcher-redis redis-cli
   
   # If using local Redis:
   redis-cli
   ```

2. Publish a test GitHub push webhook payload:
   ```bash
   PUBLISH github-webhooks '{"ref":"refs/heads/main","repository":{"full_name":"owner/repository-name"}}'
   ```

3. You should see the dispatcher process the webhook and push the matched configuration to the pipeline queue.

4. Check the pipeline queue:
   ```bash
   LRANGE pipeline 0 -1
   ```

### Running Unit Tests

```bash
go test ./... -v
```

To run only unit tests (skip integration tests):
```bash
go test ./... -v -short
```

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
- **main.go**: Main service logic that:
  - Connects to Redis and listens for pubsub messages
  - Loads filter configuration from JSON file
  - Parses GitHub push webhook payloads
  - Matches webhooks against configured repository/branch filters
  - Pushes matched configurations to Redis queue for pipeline processing
- **Docker Compose**: Orchestrates Redis and the dispatcher service
- **Dockerfile**: Multi-stage build for creating a minimal production image
- **config.json**: Filter configuration defining which repos/branches to process

### Workflow

1. Service starts and loads filter rules from `config.json`
2. Service subscribes to Redis pubsub channel (`github-webhooks`)
3. When a GitHub push webhook is received:
   - Parse the webhook payload to extract repository and branch
   - Check if it matches any configured filter rule
   - If matched, serialize the rule configuration and push to Redis queue (`pipeline`)
4. Pipeline workers can consume from the queue to execute CI/CD commands

## Future Enhancements

- Support for additional GitHub event types (pull requests, issues, etc.)
- Webhook signature verification for security
- Add structured logging framework
- Add metrics and monitoring
- Support for more complex matching patterns (wildcards, regex)
- Dead letter queue for failed processing
