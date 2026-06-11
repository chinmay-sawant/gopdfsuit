#!/usr/bin/env bash
# Gin HTTP load test + CPU/heap pprof capture for /api/v1/generate/template-pdf.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
OUT="${REPO_ROOT}/guides/cursor/baselines/gin_pprof_runs"
DATE_TAG="$(date +%Y%m%d_%H%M%S)"
BIN="${OUT}/gopdfsuit_${DATE_TAG}"
CPU_PROF="${OUT}/cpu_gin_${DATE_TAG}.prof"
HEAP_PROF="${OUT}/heap_gin_${DATE_TAG}.prof"
K6_LOG="${OUT}/k6_gin_${DATE_TAG}.txt"
PPROF_SUMMARY="${OUT}/pprof_summary_${DATE_TAG}.txt"

LOAD_VUS="${LOAD_VUS:-48}"
PROFILE_SECONDS="${PROFILE_SECONDS:-35}"
PAYLOAD_SCENARIO="${PAYLOAD_SCENARIO:-tagged_ecdsa}"
BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
GOMAXPROCS_VAL="${GOMAXPROCS:-24}"
MAX_CONCURRENT_VAL="${MAX_CONCURRENT:-48}"
BENCH_MODE_VAL="${BENCH_MODE:-1}"
GIN_FAST_API_VAL="${GIN_FAST_API:-1}"
THROUGHPUT_GATE="${THROUGHPUT_GATE:-0}"

mkdir -p "$OUT"

echo "==> Building server binary"
(cd "$REPO_ROOT" && go build -o "$BIN" ./cmd/gopdfsuit)

# Kill anything on :8080 from a prior run
if command -v fuser >/dev/null 2>&1; then
  fuser -k 8080/tcp 2>/dev/null || true
fi
sleep 1

echo "==> Starting server (GOMAXPROCS=$GOMAXPROCS_VAL, MAX_CONCURRENT=$MAX_CONCURRENT_VAL, BENCH_MODE=$BENCH_MODE_VAL, GIN_FAST_API=$GIN_FAST_API_VAL)"
export GOMAXPROCS="$GOMAXPROCS_VAL"
export MAX_CONCURRENT="$MAX_CONCURRENT_VAL"
export BENCH_MODE="$BENCH_MODE_VAL"
export GIN_FAST_API="$GIN_FAST_API_VAL"
# Never ENABLE_PROFILING=1 during k6 (heap instrumentation skews results).
unset ENABLE_PROFILING
# Profiling routes are always registered; do not set ENABLE_PROFILING=1 (heap instrumentation overhead).
"$BIN" > "${OUT}/server_${DATE_TAG}.log" 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true; wait "$SERVER_PID" 2>/dev/null || true' EXIT

for i in $(seq 1 30); do
  if curl -sf "${BASE_URL}/gopdfsuit" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "Server exited early; log:"
    tail -50 "${OUT}/server_${DATE_TAG}.log" || true
    exit 1
  fi
  sleep 0.5
done

echo "==> CPU profile (${PROFILE_SECONDS}s) during k6 steady load"
curl -sf -o "$CPU_PROF" "${BASE_URL}/debug/pprof/profile?seconds=${PROFILE_SECONDS}" &
CURL_PID=$!

sleep 2

echo "==> k6 load (${LOAD_VUS} VUs, scenario=${PAYLOAD_SCENARIO})"
(
  cd "${REPO_ROOT}/test/generate_template-pdf"
  k6 run load_test_pprof.js \
    -e "BASE_URL=${BASE_URL}" \
    -e "LOAD_VUS=${LOAD_VUS}" \
    -e "PROFILE_SECONDS=${PROFILE_SECONDS}" \
    -e "PAYLOAD_SCENARIO=${PAYLOAD_SCENARIO}" \
    -e "SKIP_SMOKE=1" \
    -e "THROUGHPUT_GATE=${THROUGHPUT_GATE}"
) 2>&1 | tee "$K6_LOG"

wait "$CURL_PID"

echo "==> Heap profile (post-load)"
curl -sf -o "$HEAP_PROF" "${BASE_URL}/debug/pprof/heap"

echo "==> pprof text summary"
{
  echo "# Gin pprof summary — ${DATE_TAG}"
  echo "binary: $BIN"
  echo "k6_log: $K6_LOG"
  echo "GOMAXPROCS=$GOMAXPROCS_VAL MAX_CONCURRENT=$MAX_CONCURRENT_VAL BENCH_MODE=$BENCH_MODE_VAL GIN_FAST_API=$GIN_FAST_API_VAL LOAD_VUS=$LOAD_VUS PROFILE_SECONDS=$PROFILE_SECONDS PAYLOAD_SCENARIO=$PAYLOAD_SCENARIO"
  echo ""
  echo "## k6 throughput (grep)"
  grep -E 'http_reqs|http_req_duration|iterations|pdf_generation_time' "$K6_LOG" | tail -20 || true
  echo ""
  echo "## CPU top (flat)"
  go tool pprof -top -nodecount=25 "$BIN" "$CPU_PROF" 2>/dev/null || true
  echo ""
  echo "## CPU top (cum)"
  go tool pprof -top -cum -nodecount=25 "$BIN" "$CPU_PROF" 2>/dev/null || true
  echo ""
  echo "## Heap top (inuse_space)"
  go tool pprof -top -inuse_space -nodecount=20 "$BIN" "$HEAP_PROF" 2>/dev/null || true
  echo ""
  echo "## Heap top (alloc_space)"
  go tool pprof -top -alloc_space -nodecount=20 "$BIN" "$HEAP_PROF" 2>/dev/null || true
} | tee "$PPROF_SUMMARY"

echo ""
echo "Artifacts:"
echo "  CPU:     $CPU_PROF"
echo "  Heap:    $HEAP_PROF"
echo "  Summary: $PPROF_SUMMARY"
echo "  k6:      $K6_LOG"