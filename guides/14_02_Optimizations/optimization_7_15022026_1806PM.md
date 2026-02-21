# API Throughput Optimization: 80 → 300-400 ops/sec

**Current state:** Pure Go library hits ~1700 ops/sec on 24 cores, but the HTTP API is stuck at ~80 ops/sec.  
**Goal:** Reach 300-400 ops/sec on the API.

## Root Cause Analysis

The **21x gap** (1700 vs 80) is far too large to be explained by HTTP serialization alone. After code review, these are the identified bottlenecks:

### 1. Profiling Overhead (mem.prof writes)

Memory profiling is always enabled unless `DISABLE_PROFILING` is set. While the `WriteHeapProfile` only runs at shutdown, the runtime overhead of heap profiling instrumentation is always active and adds per-allocation cost.

### 2. Semaphore Set to 100 (Over-Provisioned)

`maxConcurrent = 100` on a 24-core machine means up to 100 goroutines compete for CPU, causing excessive context switching. This is the single biggest bottleneck. PDF generation is **CPU-bound**, so the optimal concurrency is `runtime.NumCPU()` (24).

### 3. `gin.Recovery()` Stack Trace Overhead

`gin.Recovery()` captures full stack traces on every panic. Even without panics, it wraps the handler in a `defer/recover` that adds overhead per request.

### 4. Middleware Stack on Hot Path (CORS + Auth)

Every API request passes through CORS and GoogleAuth middleware. The auth middleware calls `IsCloudRun()` which reads environment variables (`os.Getenv`) on every request.

### 5. k6 Tests VU Counts Too High

Spike test ramps to 200 VUs, which on a local machine creates artificial contention. All tests need VU reduction to 100.

---

## Proposed Changes

### Server Configuration (`cmd/gopdfsuit`)

#### [MODIFY] `main.go`

1. **Set semaphore to `runtime.NumCPU()`** instead of hardcoded 100
   - This is the most impactful change — prevents 100 goroutines fighting for 24 cores
   - CPU-bound workloads perform best when concurrency matches available cores

2. **Default profiling to DISABLED** — flip the logic so profiling is opt-in via `ENABLE_PROFILING=1`
   - Removes heap profiling instrumentation overhead during benchmarks

3. **Replace `gin.Recovery()` with a lightweight custom recovery** middleware
   - Only captures stack traces when a panic actually occurs (lazy evaluation)
   - Avoids per-request overhead of the default `gin.Recovery()`

---

### Middleware Optimization (`internal/middleware`)

#### [MODIFY] `auth.go`

4. **Cache `IsCloudRun()` result** — evaluate once at startup, store in a package-level variable
   - Eliminates 2x `os.Getenv` calls per request

---

### Handler Optimization (`internal/handlers`)

#### [MODIFY] `handlers.go`

5. **Use `io.ReadAll` instead of `c.GetRawData()`** — the `GetRawData()` method does an extra copy internally. Using `io.ReadAll(c.Request.Body)` directly avoids that.

---

### k6 Test Updates (`test/generate_template-pdf`)

#### [MODIFY] `spike_test.js`

6. Cap spike VUs at 100 (currently 200)

#### [MODIFY] `soak_test.js`

7. Lower soak VUs (currently 10 but `sleep(2)` limits throughput) — remove sleep to allow real throughput measurement

#### [MODIFY] `load_test.js`

8. Cap load test ramp to 100 VUs (currently 5, which is fine — leave as-is or bump slightly for meaningful load)

---

## Expected Impact

| Change                           | Estimated Impact     |
| -------------------------------- | -------------------- |
| Semaphore CPU-match (100→24)     | **+150-200 ops/sec** |
| Disable profiling                | **+20-40 ops/sec**   |
| Recovery middleware optimization | **+10-20 ops/sec**   |
| Auth caching                     | **+5-10 ops/sec**    |
| Handler body read optimization   | **+5-10 ops/sec**    |

Conservative estimate: **300-400 ops/sec** achievable.

---

## Verification Plan

### Automated Tests

1. Run existing integration tests to verify nothing is broken:

   ```bash
   go test ./...
   ```

2. Run the k6 smoke test as a quick validation (server must be running):

   ```bash
   # Terminal 1: Start server
   DISABLE_PROFILING=1 go run cmd/gopdfsuit/main.go

   # Terminal 2: Run smoke test
   cd test/generate_template-pdf && k6 run smoke_test.js
   ```

3. Run the k6 load test to measure throughput:

   ```bash
   cd test/generate_template-pdf && k6 run load_test.js
   ```

   **Success criteria:** `http_reqs` rate should show 300+ requests/second during the sustained load phase.

### Manual Verification

- After starting the server, manually verify it still responds correctly:
  ```bash
  curl -X POST http://localhost:8080/api/v1/generate/template-pdf \
    -H "Content-Type: application/json" \
    -d '{"config":{"page":"A4"},"title":{"text":"Test"},"elements":[]}' \
    -o /dev/null -w "%{http_code}\n"
  ```
  Expected: `200`
