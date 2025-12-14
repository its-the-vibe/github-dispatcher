# Copilot Instructions for github-dispatcher

## Project Overview

This is a Go-based service that receives GitHub webhook notifications from a Redis pubsub channel and dispatches CI/CD operations. The service is designed to be lightweight, containerized, and production-ready.

## Architecture

- **Language**: Go 1.24 or later
- **Main Dependencies**: 
  - `github.com/redis/go-redis/v9` - Redis client for pubsub functionality
- **Deployment**: Docker and Docker Compose
- **Core Components**:
  - `main.go` - Main service logic with Redis pubsub subscription
  - `main_test.go` - Unit tests for configuration loading
  - `Dockerfile` - Multi-stage build for minimal production images
  - `docker-compose.yaml` - Orchestration for Redis and dispatcher service

## Development Setup

### Prerequisites
- Go 1.24 or later
- Docker and Docker Compose (optional, for containerized deployment)
- Redis server (can use Docker Compose)

### Installing Dependencies
```bash
go mod download
```

### Running Locally
```bash
go run main.go
```

### Running with Docker Compose
```bash
docker-compose up --build
```

## Building and Testing

### Build Command
```bash
go build -o github-dispatcher .
```

### Test Command
```bash
go test ./...
```

### Test with Verbose Output
```bash
go test ./... -v
```

## Code Style and Conventions

### General Go Conventions
- Follow standard Go conventions and idioms
- Use `go fmt` for code formatting
- Use descriptive variable and function names
- Keep functions small and focused

### Configuration
- All configuration is done via environment variables
- Use the `getEnv()` helper function with sensible defaults
- Configuration struct is defined in `Config` type

### Environment Variables
- `REDIS_HOST` - Redis server hostname (default: `localhost`)
- `REDIS_PORT` - Redis server port (default: `6379`)
- `REDIS_CHANNEL` - Redis pubsub channel name (default: `github-webhooks`)

### Error Handling
- Use `log.Fatalf()` for critical errors that should terminate the service
- Log errors with descriptive messages using `log.Printf()`

### Logging
- Use the standard `log` package
- Include context in log messages (e.g., channel name, configuration values)
- Log key lifecycle events (startup, connection success, shutdown)

## Testing Practices

### Test Organization
- Tests are in `*_test.go` files alongside the code they test
- Use table-driven tests where appropriate
- Test both default and custom configuration values

### Testing Configuration
- Always clean up environment variables in tests using `os.Unsetenv()`
- Test both presence and absence of environment variables

### Running Tests Locally
Run tests to verify changes:
```bash
go test ./... -v
```

## Docker and Containerization

### Dockerfile
- Uses multi-stage build for minimal image size
- Production image should be based on minimal base images
- Binary is built in build stage and copied to runtime stage

### Docker Compose
- Includes Redis service for local development
- Dispatcher service depends on Redis being healthy
- Use environment variables for configuration

## Project Structure

```
github-dispatcher/
├── .github/              # GitHub-specific configuration
│   └── copilot-instructions.md
├── main.go              # Main service implementation
├── main_test.go         # Unit tests
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Dockerfile           # Container image definition
├── docker-compose.yaml  # Multi-container orchestration
├── .env.example         # Example environment configuration
└── README.md            # Project documentation
```

## Common Tasks

### Adding New Configuration Options
1. Add field to `Config` struct in `main.go`
2. Update `loadConfig()` function with `getEnv()` call
3. Add tests in `main_test.go` for the new configuration
4. Update `.env.example` and README.md with the new variable

### Testing Message Processing
1. Start services: `docker-compose up --build`
2. In another terminal, connect to Redis: `docker exec -it github-dispatcher-redis redis-cli`
3. Publish test message: `PUBLISH github-webhooks "Test webhook message"`
4. Verify message appears in dispatcher logs

## Future Enhancements (Planned)

- Parse and validate GitHub webhook payloads
- Trigger CI/CD operations based on webhook events
- Add structured logging framework
- Add metrics and monitoring
- Implement webhook signature verification
- Add more comprehensive error handling

## Important Notes

- This service is designed for graceful shutdown (handles SIGINT and SIGTERM)
- Redis connection is tested on startup before subscribing
- The service uses blocking operations in the main loop with signal handling
- All Redis operations use context for proper cancellation support
