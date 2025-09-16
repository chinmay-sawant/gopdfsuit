# Build from the root directory using: docker build -f dockerfolder/Dockerfile -t gopdfsuit .

# --- Build stage ---
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


# --- Final stage ---
# Use Debian 12 for a native (non-snap) Chromium package
FROM debian:12-slim

ARG VERSION=latest
LABEL version=$VERSION

# Set DEBIAN_FRONTEND to noninteractive to prevent apt-get from prompting for input
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies for headless Chromium, install Chromium itself, and clean up.
# We need fonts and other libraries for it to render pages correctly.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    chromium \
    # Add common dependencies for headless browsers to prevent rendering issues
    libxss1 \
    libasound2 \
    libnspr4 \
    libnss3 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libx11-xcb1 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    fonts-liberation \
    ca-certificates \
    && echo "Cleaning up..." \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app/

# Copy the binary from builder stage
COPY --from=builder /app/gopdfsuit .

# Copy the web folder
COPY --from=builder /app/docs ./docs

# Copy LICENSE
COPY --from=builder /app/LICENSE .

# Copy README.md
COPY --from=builder /app/README.md .

# Copy temp_multiplepage.json
COPY --from=builder /app/temp_multiplepage.json .

# Expose port if the app serves on a port (assuming 8080, adjust if needed)
EXPOSE 8080

# Run the binary.
# NOTE: Your Go application likely needs to launch Chromium with the '--no-sandbox' flag
# when running as the root user inside a Docker container.
CMD ["./gopdfsuit"]