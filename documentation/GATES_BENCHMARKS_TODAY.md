# Gates benchmarks today

Date: 2026-09-04. Branch: feat/builder-snippets. Source of truth is lowercase makefile. guides/MAKEFILE.md is a frozen archive per its header.

Note: test/integration_coverage_test.go does not exist. Actual test files are integration_test.go, integration_misc_test.go, integration_xfdf_test.go, integration_redact_test.go, integration_financial_report_test.go, zerodha_compliance_test.go, benchmark_handlers_test.go, helpers_test.go.

## Make gates

- fmt at makefile:206 runs go fmt ./...
- vet at 209 runs go vet ./...
- lint at 215 runs golangci-lint run -E revive,gocritic,gocyclo,goconst ./... plus cd frontend plus npm run lint
- test full at 82 runs test-go plus test-python plus test-verify
- test-go at 84 runs test-unit then test-integration
- test-unit at 92 runs go test except ./test package
- test-integration at 99 runs sequential test-integration-suite plus test-integration-zerodha
- test-integration-suite at 107 runs go test -p 1 -parallel 1 with TestIntegrationSuite in ./test
- test-integration-zerodha at 110 runs TestZerodhaPDFCompliance in ./test
- test-python at 113 runs python3 -m pytest tests in bindings/python
- test-verify-pdfs at 118 runs bash test/verify_pdfs.sh, alias test-verify at 121
- test-fast tier1 at 123 runs test-unit SHORT=1 plus test-python
- test-prepush tier2 at 127 runs test-fast plus test-integration
- test-nightly tier4 at 129 runs test plus test-scan-pdfs-compliance plus test-race
- test-race at 151 runs go test -race on internal/pdf and internal/handlers
- test-schema at 139 runs frontend npm test:schema
- test-zerodha-compliance at 166 runs verify_pdfs.sh --zerodha-only
- build at 19 runs test-go then go build -o bin/app ./cmd/gopdfsuit
- wasm-compress at 182 runs GOOS=js GOARCH=wasm build to frontend/public/compress.wasm plus cp wasm_exec.js plus file check
- wasm full at 191 builds frontend/public/gopdfsuit.wasm from ./cmd/wasm
- run at 199 runs test-integration plus lint plus wasm-compress then frontend build plus go run
- Defaults at 52: JOBS and P 2, TEST_PARALLEL 2, TEST_TIMEOUT 10m, TEST_GOMEMLIMIT 4GiB

## Benchmark numbers

From BENCHMARKS.md dated 2026-06-18 on feat/optimization-5.5-medium, WSL2 i7-13700HX 24 CPUs, Go 1.26.4, best of 5:

- bench-gopdflib-zerodha with PDF/A-4 plus PDF/UA-2: 6611 ops/s peak x10, mean 6203, median 6362, 7.5ms avg latency
- bench-gopdflib-zerodha-nocomply PDF 2.0: 37853 ops/s peak, mean 34035
- bench-gopdfsuit-zerodha retail: 6146 ops/s
- bench-gopdflib-data 2000 rows: 288 ops/s
- bench-k6 48VU 35s weighted: 1333 req/s, light 24VU 15s: 1177 req/s, retail: 7515 req/s, Gotenberg 16.1 req/s about 83x slower
- Handler BenchmarkGenerateTemplatePDF_FinancialReport: 55528 ns/op, parallel 54814 ns/op
- GoPDFKit table_180_rows: 47707 vs 11883 pdf/s, 4.0x
- bench-pypdfsuit-zerodha: 937 peak and 916 mean, nocomply 1284 peak and 1242 mean

Older layer doc INTEGRATION_AND_BENCHMARK_TESTS.md from 2026-06-11: handler serial 1045205 ns/op, parallel 133124 ns/op, gopdflib Zerodha 2953 peak and 2646 mean over 30 runs, k6 859 peak and 825 avg over 5 runs, retail gate 3965 req/s, Gotenberg 10.3 req/s.

## Caching and pools

- templatePDFPool plus resetTemplate and acquire and release at handlers/decode.go:10-37. sync.Pool of PDFTemplate cleared through ResetForReuse.
- bodyBufPool 64 KiB new at handlers/json_decode.go:24. Pooled only when contentLength under 512 KiB. Discard when cap above 128 KiB at putBodyBuf:91.
- hftBodyBufPool 2 MiB new at json_decode.go:30. HFT path under 8 MiB. Discard when cap above 8 MiB at putHftBodyBuf:101.
- pdfBufferPoolSmall and Large at pdf/generator.go:91-105. Retail 96 KiB, active 128 KiB, HFT 2816 KiB, max pooled 3 MiB.
- xrefOffsetsSlicePool 2048 ints and usedXrefIDsPool 64 ints at generator.go:154. Drop when cap above 64K at release:195.
- ZlibWriterPool and CompressBufPool at pdf/font/compression.go:13 and 22.
- Page compress cache sharded at pdf/font/compress_cache.go:20-41 and 111-135. Max 2048 entries, fingerprint cap 32K, per shard Clear on overflow.
- Font subset cache at pdf/font/subset_cache.go:17-34 and 62-79. Max 1024 entries keyed by FNV PostScriptName plus sorted glyphs.
- Image cache plus MRU slot at pdf/image.go:37-93. Max 256 entries keyed by FNV base64 plus single slot MRU.
- rgbDataPool at image.go:25. 1 MiB new, drop above 8 MiB.
- structElemPool, arenaSlabPool, structKidsSlicePool at structure.go:174, 180, 238.
- pagemanager pool at pagemanager.go:24, registryClonePool at font/registry.go:247, signature pools at signature/signature.go:104.

Drift note: CACHING_AND_MEMORY_LIFECYCLE.md:60-88 calls subset and image caches unbounded with no expiry. Current code is bounded at 1024 and 256 with clear APIs.

Code:

```go
const maxPageCompressCacheEntries = 2048
const maxFingerprintCachedContentLen = 32 * 1024
```

```go
func putBodyBuf(bufPtr *[]byte, buf []byte) {
	// Never return large backing arrays to the pool - keeps heap bounded under 48 workers.
	if cap(buf) > 128<<10 {
		*bufPtr = make([]byte, 0, 64*1024)
	} else {
		*bufPtr = buf[:0]
	}
	bodyBufPool.Put(bufPtr)
}
```

Build gate:

```
build: test-go
	mkdir -p bin
	go build -o bin/app ./cmd/gopdfsuit
```

## Reproduce

- make test-fast for tier1
- make test-prepush for tier2
- make test-verify-pdfs for manifest gate
- make bench-smoke and make bench-help for bench entry
- Schema parity pins: TestHandlerInputJSONParity and TestEngineOptionFieldParity at models/schema_parity_test.go:70 and 88
- Handler bench: BenchmarkGenerateTemplatePDF_FinancialReport at test/benchmark_handlers_test.go:34
