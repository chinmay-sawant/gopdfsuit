# Handlers fast path and envelope

Date: 2026-09-04. Branch: feat/builder-snippets.

## Routes

All under v1 group /api/v1 with GoogleAuthMiddleware at handlers.go:89-139. GIN_FAST_API benchmark path skips CORS only.

- POST /api/v1/generate/template-pdf to handleGenerateTemplatePDF at generate.go:10
- POST /api/v1/fill to handleFillPDF at fill.go:11
- POST /api/v1/merge to handleMergePDFs at merge_split.go:17
- POST /api/v1/split to handlerSplitPDF at merge_split.go:56
- POST /api/v1/compress to handleCompressPDF at compress.go:14
- GET /api/v1/template-data to handleGetTemplateData
- GET /api/v1/fonts to handleGetFonts at fonts.go:13, POST /api/v1/fonts to handleUploadFont at fonts.go:23
- POST /api/v1/htmltopdf to handleHTMLToPDF at html.go:32, POST /api/v1/htmltoimage to handleHTMLToImage at html.go:62
- POST /api/v1/redact/page-info at redact.go:72, text-positions at 104, capabilities at 88, apply at 131, search at 170

Success types: PDF ops return application/pdf. Split multi output returns application/zip at merge_split.go:122. htmltoimage returns image png or jpeg. template-data returns application/json.

## Fast vs normal generate

Decode at generate.go:10-29, decode.go:45-47, json_decode.go:49-89, hft_decode.go:42-51:

- acquireTemplate pulls PDFTemplate from templatePDFPool at decode.go:27 with ResetForReuse before and after.
- PreallocForDecode uses ContentLength and X-Payload-Tier to size arrays.
- Body capped by http.MaxBytesReader with maxTemplateJSONBody 8 MiB at json_decode.go:19 and generate.go:20. Oversize calls abortError 413 template too large.
- decodeTemplateJSON tiers: tier hft with known length under 8 MiB reads once into hftBodyBufPool 2 MiB seed plus decodeHFTPayload with JIT row unmarshal through fillTableRowsFast. Unknown or large HFT uses stream decoder. Retail with known length under 512 KiB uses bodyBufPool 64 KiB seed plus sonic.Unmarshal. Larger or unknown uses stream decode with no copy.
- WarmJSONDecode pretouches PDFTemplate schema plus decoder plus HFT path at json_decode.go:39-47.

Render at generate.go:37-58 and services.go:62-79:

- Fast borrowed path with zero extra copy: when pdfService matches FastGenerateService, call GenerateTemplatePDFBorrowed, defer Release, write bytes with status 200.
- Normal fallback: GenerateTemplatePDF returns bytes, sent with c.Data 200.
- Render failure maps to abortError 500 PDF generation failed. Decode failure maps to abortError 400 invalid template data.
- Proven by generate_fast_test.go:19 TestGenerateBorrowedSeam with FastMockPDFService wrapping gomock MockPDFService. Mock lives at mocks/mock_fast_generate.go:13.

## Error envelope

Defined at request.go:89-127, taxonomy in pkg/gopdflib/errors.go.

Wire shape: {"code": "<code>", "message": "<msg>", "error": "<msg>"} where error is a legacy alias of message at request.go:91-93.

Codes: invalid_input for 422, limit_exceeded for 413, upstream for 502, internal for 500, plus transport codes through CodeForStatus for direct abortError like 400, 403, 404, 413.

- abortError at 97: status plus CodeForStatus code, caller message verbatim.
- abortPDFError at 111: logs backend detail server side, classifies through pdfService.ClassifyError to pdfErrorStatus at 69. Order is ErrInternal first, then gopdflib.CodeOf sentinels, then font upstream errors to 502, else ClassifyMessage substring fallback. Client messages stay generic: 422 invalid PDF input, 413 pdf exceeds maximum size, 502 upstream dependency failed, else 500 plus fallback like redaction failed or PDF processing failed.
- Pinned bodies in error_envelope_test.go: 500 redaction failed at 70, 422 invalid PDF input at 101, 502 upstream dependency failed at 118. pprof deny reuses same errorBody at request.go:186.

Upload policy: UploadLimit and ReadUpload on PDFService at services.go:40-49, caps in request.go:21-30. readSingleUpload and readUploadData at 200 plus 231 return 400 for missing or empty and 413 with overLimitMessage, then abort.

## Per op notes

- merge_split.go:17 merge, 56 split with single PDF out vs splits.zip.
- compress.go:14 level through gopdflib.ParseCompressLevel, default Medium.
- fill.go:11 accepts file or pdf_bytes plus xfdf_bytes raw fallback with limit check.
- redact.go:131 apply returns PDF plus X-Redaction-Report header. Parser parseRedactApply at redact_parse.go:28 handles blocks, textSearch, ocr plus legacy redactions and text.
- html.go:32 and 62 guard with validateFetchURL, svg rejected 400.

## Snippets

Generate request from generate_fast_test.go:41-44:

```json
// POST /api/v1/generate/template-pdf
// Optional header: X-Payload-Tier: hft
{"config":{"page":"A4","pageAlignment":1},"title":{"props":"a","text":"b"},"footer":{"font":"a","text":"b"}}
// 200 application/pdf
// 400 {"code":"invalid_input","message":"invalid template data","error":"invalid template data"}
// 413 {"code":"limit_exceeded","message":"template too large","error":"template too large"}
```

Merge invalid input from error_envelope_test.go:91-105:

```json
// POST /api/v1/merge with multipart pdf fields
// backend wraps gopdflib.ErrInvalidInput
// 422 {"code":"invalid_input","error":"invalid PDF input","message":"invalid PDF input"}
```

Redact apply from redact.go:131 plus redact_parse.go:28:

```json
// POST /api/v1/redact/apply with pdf file plus text fields
// fields: mode, password, blocks, textSearch, ocr, redactions legacy, text legacy
// 200 application/pdf plus header X-Redaction-Report
// 400 {"code":"invalid_input","message":"invalid blocks json","error":"invalid blocks json"}
// 500 {"code":"internal","error":"redaction failed","message":"redaction failed"}
```

Upstream failure from error_envelope_test.go:107-122:

```json
// backend returns font.HTTPStatusError 503
// 502 {"code":"upstream","error":"upstream dependency failed","message":"upstream dependency failed"}
```

Key refs: request.go:69 pdfErrorStatus, 91 errorBody, 97 abortError, 111 abortPDFError. services.go:33 PDFService, 67 FastGenerateService, 169 SetPDFService. decode.go:27 acquireTemplate, 45 decodeTemplate. json_decode.go:49 decodeTemplateJSON. hft_decode.go:42 decodeHFTPayload.
