# --- Stage 1: Build the Go application ---
FROM golang:1.24-bookworm AS builder
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/ ./docs/
COPY README.md LICENSE ./

# Strip debug symbols (-s -w) to reduce binary size by ~25%
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/gopdfsuit/main.go

# --- Stage 2: Optimized Runtime ---
# This image contains a stripped-down Chromium optimized for automation
# Uncompressed size is approx ~300-350MB
FROM chromedp/headless-shell:latest
RUN rm -rf /usr/share/doc /usr/share/man /usr/share/locale
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Install dumb-init to prevent zombie processes (common issue with Chrome in Docker)
RUN apt-get update && apt-get install -y --no-install-recommends dumb-init \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs

# The headless-shell binary is located here
ENV CHROME_PATH="/headless-shell/headless-shell"
ENV GOPDFSUIT_ROOT="/app"

EXPOSE 8080

ENTRYPOINT ["dumb-init", "--"]
CMD ["./server"]