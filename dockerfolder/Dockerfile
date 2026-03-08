# Build from the root directory using: docker build -f dockerfolder/Dockerfile -t gopdfsuit .

# --- Build stage ---
FROM golang:1.24.11-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
	rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy only the sources needed to build the server.
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY typstsyntax/ ./typstsyntax/
COPY docs/ ./docs/
COPY README.md LICENSE temp_multiplepage.json ./

# Strip debug symbols and build a static binary to reduce image size.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o gopdfsuit ./cmd/gopdfsuit


# --- Final stage ---
# Use the smaller headless-shell runtime instead of installing full Chromium.
FROM chromedp/headless-shell:latest

ARG VERSION=latest
LABEL version=$VERSION

RUN rm -rf /usr/share/doc /usr/share/man /usr/share/locale && \
	apt-get update && apt-get install -y --no-install-recommends dumb-init && \
	apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app/

# Copy runtime dependencies.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/gopdfsuit .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/LICENSE .
COPY --from=builder /app/README.md .
COPY --from=builder /app/temp_multiplepage.json .

ENV CHROME_PATH=/headless-shell/headless-shell
ENV GOPDFSUIT_ROOT=/app

EXPOSE 8080

ENTRYPOINT ["dumb-init", "--"]
CMD ["./gopdfsuit"]