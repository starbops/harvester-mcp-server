FROM docker.io/library/golang:1.23-alpine AS builder

WORKDIR /workspace

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o harvester-mcp-server ./cmd/harvester-mcp-server

# Create a minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /workspace/harvester-mcp-server /app/harvester-mcp-server

# Set the entry point
ENTRYPOINT ["/app/harvester-mcp-server"]