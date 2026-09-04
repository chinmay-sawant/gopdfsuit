# Auth and request limits today

Date: 2026-09-04. Commit eae7a7a plus router and middleware state on feat/builder-snippets.

Truth lives in documentation per plans/adr-2026-09-04-doc-homes.md. Never hand edit docs output.

## Env matrix

Backend:

- REQUIRE_AUTH=1 forces enforcement outside Cloud Run. Else open by default.
- K_SERVICE and K_REVISION auto set by Cloud Run and trigger enforcement.
- Audience precedence for idtoken.Validate: GOOGLE_OAUTH_AUDIENCE else GOOGLE_CLIENT_ID else CLOUD_RUN_SERVICE_URL.
- GIN_FAST_API=1 sets RoutePolicy.EnableCORS false only. It never skips auth.
- MAX_CONCURRENT explicit wins when integer above 0.
- BENCH_MODE=1 fallback is NumCPU times 2, capped at 48, min 1. Else NumCPU.
- ENABLE_PROFILING=1 dumps heap profile to /tmp/mem.prof. Opt in only.
- GOPDFSUIT_ROOT overrides project root for docs static serving.
- Server listens on :8080.

Frontend:

- VITE_GOOGLE_CLIENT_ID required.
- VITE_IS_CLOUD_RUN true or VITE_ENVIRONMENT cloudrun enables AuthGuard.
- VITE_API_URL, VITE_CLOUD_RUN_URL.

## Auth flow

All /api/v1/* routes sit in one group at handlers.go:120-138 with GoogleAuthMiddleware. Protected: generate template-pdf, fill, merge, split, compress, template-data, fonts GET and POST, htmltopdf, htmltoimage, redact page-info plus text-positions plus capabilities plus apply plus search.

Rules from internal/middleware/auth.go:

- IsCloudRun at 21 checks K_SERVICE or K_REVISION.
- authEnforced at 33 is true when REQUIRE_AUTH is 1 or IsCloudRun.
- resolveAudience at 44 applies precedence above.
- Open by default: when not enforced the middleware calls Next with no checks at 65.
- OPTIONS preflight skipped even when enforced at 74.
- Missing header returns 401 Authorization header required. Bad scheme returns 401 with Bearer hint. Validate fail returns 401 authentication failed. Success stores user_email, user_name, user_picture, user_sub at 114.
- OptionalAuthMiddleware only checks on Cloud Run and never aborts.
- GIN_FAST_API scope: skips CORSMiddleware only. Template-pdf stays inside v1 auth group.
- /debug/pprof/* is separate, localhost only through isLoopbackPeer at request.go:244, else 403. It ignores X-Forwarded-For.

## Limits

From request.go:21-30, json_decode.go:14-20, compress/limits.go:3-16, router.go:52-66 and 139-156:

- Single PDF upload maxPDFBytes = compress.MaxInputBytes = 32 MiB.
- MaxInflateBytes = 48 MiB per Flate stream.
- MaxImagePixels = 16M, MaxImageEdge = 8192, MaxObjects = 50000, maxAllowedDim = 4096.
- XFDF max 8 MiB. Font max 10 MiB.
- HTML convert JSON max 2 MiB through decodeJSONBody plus http.MaxBytesReader, 413 on overflow.
- Template-pdf JSON max 8 MiB through http.MaxBytesReader in handleGenerateTemplatePDF, 413 template too large.
- Decode tiers: pooled decode 512 KiB, HFT encode 8 MiB.
- Concurrency: concurrencyLimiter semaphore, non blocking try acquire, full returns 429 server busy plus Retry-After 1.
- Timeouts: Read 30s, Write 60s, Shutdown 15s.
- SSRF: validateFetchURL allows only http and https, blocks localhost, loopback, private, link local, multicast, checks DNS resolved IPs. Blocked returns 403 url target is not allowed.
- Status map: 413 over cap, 400 invalid template or empty upload, 422 invalid PDF input, 502 upstream font, 500 fallback, 429 busy, 401 auth.

## Cookbook

Public by default vs opt in enforcement:

```bash
# Public by default: 200 with no token
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/api/v1/fonts
# Opt-in enforcement: 401 with no token
REQUIRE_AUTH=1 go run ./cmd/gopdfsuit &
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/api/v1/fonts
kill %1
```

Authenticated request:

```bash
curl -s http://localhost:8080/api/v1/fonts \
  -H "Authorization: Bearer $GOOGLE_ID_TOKEN"
```

Go composition:

```go
cfg := handlers.ResolveServerConfig()
router := handlers.NewRouter(cfg)
srv := handlers.NewServer(cfg, router)
```

See AUTHENTICATION.md:103-129 for REQUIRE_AUTH outside Cloud Run and GIN_FAST_API never bypassing auth. See TROUBLESHOOTING_AUTH.md:63-66 for audience order. See DEPLOYMENT_CHECKLIST.md:264-273 for 413 caps.
