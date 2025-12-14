# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o github-dispatcher .

# Runtime stage
FROM scratch

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/github-dispatcher .

ENTRYPOINT ["./github-dispatcher"]
