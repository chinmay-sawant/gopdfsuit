# Build from the root directory using: docker build -f dockerfolder/Dockerfile -t gopdfsuit .
# Build stage
FROM golang:1.23.0-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -o gopdfsuit ./cmd/gopdfsuit

# Final stage
FROM alpine:latest

ARG VERSION=latest
LABEL version=$VERSION

# Install ca-certificates for HTTPS requests if needed
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/gopdfsuit .

# Copy the web folder
COPY --from=builder /app/web ./web

# Copy LICENSE
COPY --from=builder /app/LICENSE .

# Copy README.md
COPY --from=builder /app/README.md .

# Copy temp_mutliplepage.json
COPY --from=builder /app/temp_mutliplepage.json .

# Expose port if the app serves on a port (assuming 8080, adjust if needed)
EXPOSE 8080

# Run the binary
CMD ["./gopdfsuit"]
