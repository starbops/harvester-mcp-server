FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o harvester-mcp-server ./cmd/harvester-mcp-server

# Create a minimal runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/harvester-mcp-server .

# Set the entry point
ENTRYPOINT ["/app/harvester-mcp-server"] 