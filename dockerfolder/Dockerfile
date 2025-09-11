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

# --- Final stage ---
# Use the Ubuntu 18.04 (Bionic) base image, which is the exact environment
# the pre-compiled wkhtmltopdf binary was built for. This guarantees
# that all dependencies like libssl1.1 will be available.
FROM ubuntu:18.04

ARG VERSION=latest
LABEL version=$VERSION

# Set DEBIAN_FRONTEND to noninteractive to prevent apt-get from prompting for input
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies, download wkhtmltopdf, install it, and clean up in a single layer.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    wget \
    ca-certificates \
    xvfb \
    fontconfig \
    libxrender1 \
    xfonts-75dpi \
    xfonts-base \
    && echo "Downloading and installing wkhtmltopdf..." \
    && wget -q https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6-1/wkhtmltox_0.12.6-1.bionic_amd64.deb -O /tmp/wkhtmltox.deb \
    # apt will now correctly find all dependencies (like libssl1.1) from the Ubuntu 18.04 repos.
    && apt-get install -y /tmp/wkhtmltox.deb \
    && echo "Cleaning up..." \
    && rm /tmp/wkhtmltox.deb \
    && apt-get purge -y --auto-remove wget \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Verify wkhtmltopdf installation
RUN wkhtmltopdf --version && wkhtmltoimage --version && echo "✅ wkhtmltopdf installation successful"

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