# PDF compress — JavaScript / WASM

Same engine as `gopdflib.CompressPDF`, compiled to WebAssembly (`GOOS=js GOARCH=wasm`). The PDF stays in the process. This sample does **not** call `POST /api/v1/compress`.

The Go library sample is [`sampledata/compress`](../compress). Limits live in `internal/pdf/compress/limits.go` and apply to the library, the HTTP handler, and this WASM path.

## Run

```bash
make wasm-compress                 # from repo root, once
node sampledata/compress-js/run.mjs
```

Writes `report_js_level_1.pdf` (Light), `report_js_level_2.pdf` (Medium), `report_js_level_3.pdf` (Heavy).

Browser (file stays in the tab):

```bash
cd sampledata/compress-js && python3 -m http.server
# open http://localhost:8000
```

```js
import { compressPDF } from './compress.js'

const out = await compressPDF(uint8, { level: 2 })
// 1 = light (JPEG 92), 2 = medium (75), 3 = heavy (50)
```

## What is handled

These are resource and input checks, not a full PDF sandbox. A hostile file should fail closed (error, or skip that stream) instead of growing without bound.

| Guard | Limit | Why |
|-------|-------|-----|
| Input size | 32 MiB | Stops a huge upload / `Uint8Array` from allocating unbounded WASM/Go memory. Checked in JS, WASM (`byteLength` + `CopyBytesToGo`), `CompressPDF`, and the HTTP handler (`LimitReader`). |
| Flate inflate | 48 MiB per stream | Zip bomb: a tiny Flate stream that expands to hundreds of MB. Same cap in `compress.decompressFlate` and `merge.decompressFlate` (object streams). |
| Image raster | 16 megapixels, edge 8192 | A `/Width 100000 /Height 100000` XObject (or a huge JPEG) would allocate gigabytes. JPEG uses `DecodeConfig` before full decode. |
| `max_image_dim` override | 4096 | HTTP `max_image_dim=999999999` cannot disable downscale or force a 1e9 target. |
| JPEG quality | 1–100 | Values above 100 are clamped. |
| Object count / object number | 50_000 | A single `50001 0 obj` used to make xref writing loop `1..N`. Object streams with `/N` above this are ignored. |
| Encrypted PDFs | rejected | `/Encrypt` in the trailer is an error; contents are not decrypted. |
| WASM argument type | `Uint8Array` only | A fake `{ byteLength: huge }` object is rejected (`CopyBytesToGo` copies 0 bytes). |
| WASM panics | recovered | Font/parse panics return `{ error: "compression failed" }` instead of killing the tab. |
| Empty / non-PDF | rejected | Missing `%PDF-` header fails before parse. |

Skipped streams (JPEG2000, JBIG2, `/DecodeParms`, SMask, odd bit depths) are left as-is rather than decoded. That is intentional: we do not implement those codecs.

## What is not handled (and is not going to be, here)

This is still a PDF rewriter running in the tab (or in the API process). Caps bound damage; they do not make compression “safe to run on untrusted PDFs with no other limits.”

| Issue | Why it remains |
|-------|----------------|
| Main-thread freeze | `goCompressPDF` is synchronous. A 32 MiB complex PDF can stall the tab until it finishes or hits a cap. A Worker/Promise split is not in this sample. |
| CPU within the caps | 32 MiB of awkward objects can still burn CPU for a long time. Regex parsing is not constant-time. |
| `goCompressPDF` is global | Any script on the same origin can call it. If the page already has XSS, the attacker already runs in the tab. |
| `POST /api/v1/compress` | Still mounted for gopdfsuit API users. The `/compress` UI and this sample do not use it. Within 32 MiB it is a server CPU surface. |
| Encrypted / JPEG2000 / JBIG2 / filter arrays | Rejected or left unchanged. Not decoded, not a silent decrypt. |
| True font subsetting | TTF glyph outlines are hollowed while **keeping original GIDs**. CFF/`FontFile3` is Flate-only. No CID remapping. |
| Malformed TTF / weird content streams | Parser is regex- and offset-based. Garbage may error, skip a stream, or (in WASM) hit the panic recover. It is not a verifying PDF reader. |
| Output may not shrink | If the rewrite is not useful, streams are skipped; some files barely change. |
| WASM memory still grows with the file | 32 MiB in plus decode buffers is expected. There is no separate WASM memory ceiling beyond the input/inflate/image caps. |
| No auth on `/compress` | By design: the file never leaves the machine. Do not treat that as “safe to expose the HTTP compress API without your own limits.” |

If you need a hard timeout or an isolated process, wrap the Go library in your own worker. This sample does not add a CLI.

## License

No Ghostscript (or other AGPL) is used. The compressor is this repo’s MIT engine (`internal/pdf/compress`), the same code as `gopdflib.CompressPDF`.

| Piece | License |
|-------|---------|
| Engine, JS wrapper, this sample | MIT (this repository) |
| `wasm_exec.js` | BSD-3-Clause, Copyright The Go Authors (file header kept) |
| `compress.wasm` | Our MIT package plus the Go WASM runtime (BSD-3-Clause), which is how `GOOS=js GOARCH=wasm` always works |

MIT and BSD-3-Clause can be shipped together. Keep the Go copyright header on `wasm_exec.js`. Do not strip it.
