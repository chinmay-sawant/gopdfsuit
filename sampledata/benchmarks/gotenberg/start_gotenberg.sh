#!/usr/bin/env bash
# Start Gotenberg for benchmark runs (Chromium HTML→PDF).
set -euo pipefail

GOTENBERG_IMAGE="${GOTENBERG_IMAGE:-gotenberg/gotenberg:8}"
GOTENBERG_PORT="${GOTENBERG_PORT:-3010}"
# Gotenberg 8.x caps chromium-max-concurrency at 6 per container.
CHROMIUM_MAX_CONCURRENCY="${CHROMIUM_MAX_CONCURRENCY:-6}"
GOTENBERG_CONTAINER="${GOTENBERG_CONTAINER:-gopdfsuit-gotenberg-bench}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required to run Gotenberg benchmarks" >&2
  exit 1
fi

docker rm -f "$GOTENBERG_CONTAINER" >/dev/null 2>&1 || true

echo "==> Starting Gotenberg ($GOTENBERG_IMAGE) on :$GOTENBERG_PORT (chromium-max-concurrency=$CHROMIUM_MAX_CONCURRENCY)"
docker run -d --name "$GOTENBERG_CONTAINER" \
  -p "127.0.0.1:${GOTENBERG_PORT}:3000" \
  "$GOTENBERG_IMAGE" \
  gotenberg \
  --chromium-auto-start=true \
  --chromium-max-concurrency="$CHROMIUM_MAX_CONCURRENCY" \
  --api-timeout=120s

BASE_URL="http://127.0.0.1:${GOTENBERG_PORT}"
for i in $(seq 1 60); do
  if curl -sf "${BASE_URL}/health" >/dev/null 2>&1; then
    echo "==> Gotenberg ready at ${BASE_URL}"
    exit 0
  fi
  sleep 0.5
done

echo "Gotenberg failed to become healthy" >&2
docker logs "$GOTENBERG_CONTAINER" 2>&1 | tail -30 || true
exit 1