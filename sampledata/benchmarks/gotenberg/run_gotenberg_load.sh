#!/usr/bin/env bash
# Gotenberg k6 load test - mirrors test/generate_template-pdf/run_gin_pprof_load.sh.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
OUT="${REPO_ROOT}/guides/cursor/baselines/gotenberg_runs"
DATE_TAG="$(date +%Y%m%d_%H%M%S)"
K6_LOG="${OUT}/k6_gotenberg_${DATE_TAG}.txt"
SUMMARY="${OUT}/summary_${DATE_TAG}.txt"

LOAD_VUS="${LOAD_VUS:-48}"
PROFILE_SECONDS="${PROFILE_SECONDS:-35}"
PAYLOAD_SCENARIO="${PAYLOAD_SCENARIO:-tagged_ecdsa}"
BASE_URL="${BASE_URL:-http://127.0.0.1:3010}"
CHROMIUM_MAX_CONCURRENCY="${CHROMIUM_MAX_CONCURRENCY:-6}"
THROUGHPUT_GATE="${THROUGHPUT_GATE:-0}"
SKIP_NETWORK_IDLE="${SKIP_NETWORK_IDLE:-1}"
MANAGE_CONTAINER="${MANAGE_CONTAINER:-1}"

mkdir -p "$OUT"

if ! command -v k6 >/dev/null 2>&1; then
  echo "k6 not found; run: make bench-k6-install" >&2
  exit 1
fi

if [[ "$MANAGE_CONTAINER" == "1" ]]; then
  bash "$SCRIPT_DIR/start_gotenberg.sh"
  trap 'docker rm -f gopdfsuit-gotenberg-bench >/dev/null 2>&1 || true' EXIT
fi

echo "==> k6 Gotenberg load (${LOAD_VUS} VUs × ${PROFILE_SECONDS}s, scenario=${PAYLOAD_SCENARIO})"
(
  cd "$SCRIPT_DIR"
  k6 run load_test_pprof.js \
    -e "BASE_URL=${BASE_URL}" \
    -e "LOAD_VUS=${LOAD_VUS}" \
    -e "PROFILE_SECONDS=${PROFILE_SECONDS}" \
    -e "PAYLOAD_SCENARIO=${PAYLOAD_SCENARIO}" \
    -e "SKIP_SMOKE=1" \
    -e "THROUGHPUT_GATE=${THROUGHPUT_GATE}" \
    -e "SKIP_NETWORK_IDLE=${SKIP_NETWORK_IDLE}"
) 2>&1 | tee "$K6_LOG"

{
  echo "# Gotenberg k6 summary - ${DATE_TAG}"
  echo "image: ${GOTENBERG_IMAGE:-gotenberg/gotenberg:8}"
  echo "base_url: ${BASE_URL}"
  echo "chromium_max_concurrency: ${CHROMIUM_MAX_CONCURRENCY}"
  echo "load_vus: ${LOAD_VUS} profile_seconds: ${PROFILE_SECONDS} scenario: ${PAYLOAD_SCENARIO}"
  echo "skip_network_idle: ${SKIP_NETWORK_IDLE}"
  echo ""
  echo "## k6 throughput"
  grep -E 'http_reqs|http_req_duration|iterations|pdf_generation_time|retail_latency|active_latency|hft_latency' "$K6_LOG" | tail -25 || true
} | tee "$SUMMARY"

echo ""
echo "Artifacts:"
echo "  k6:      $K6_LOG"
echo "  summary: $SUMMARY"