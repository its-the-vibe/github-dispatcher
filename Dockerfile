# Build stage
FROM --platform=$BUILDPLATFORM golang:1.27.0-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o github-dispatcher .

# Runtime stage
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/github-dispatcher .

USER nonroot:nonroot

ENTRYPOINT ["./github-dispatcher"]
