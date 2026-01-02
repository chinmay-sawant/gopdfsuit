# --- Stage 1: Build the Go application ---
FROM golang:1.23-bookworm AS builder
WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./
RUN go mod download

# ONLY copy the specific folders you need
# (e.g., cmd/ for entrypoint, internal/ for logic)
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/ ./docs/
COPY README.md ./
COPY LICENSE ./

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/gopdfsuit/main.go

# --- Stage 2: Optimized Runtime ---
FROM debian:bookworm-slim

# Install minimal Chrome dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget gnupg ca-certificates fonts-liberation libnss3 libatk1.0-0 \
    libatk-bridge2.0-0 libcups2 libdrm2 libxcomposite1 libxdamage1 \
    libxrandr2 libgbm1 libasound2 \
    && wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && apt-get install -y --no-install-recommends ./google-chrome-stable_current_amd64.deb \
    && rm google-chrome-stable_current_amd64.deb \
    && apt-get purge -y wget gnupg \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

# Add Chrome to the PATH
# Usually it's in /usr/bin/, but this ensures the system finds it instantly
ENV PATH="/usr/bin/google-chrome-stable:${PATH}"
# Explicit variable often used by Go libraries
ENV CHROME_PATH="/usr/bin/google-chrome-stable"

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs

# Set the project root explicitly for the application to find docs folder
ENV GOPDFSUIT_ROOT="/app"
# Expose port if the app serves on a port (assuming 8080, adjust if needed)
EXPOSE 8080

# Cloud Run expects the app to listen on $PORT
CMD ["./server"]