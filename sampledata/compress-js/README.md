# PDF compress — JavaScript / WASM

Same engine as `gopdflib.CompressPDF`, compiled to WebAssembly (`GOOS=js GOARCH=wasm`). This sample does **not** call `POST /api/v1/compress`.

The Go library sample is [`sampledata/compress`](../compress). Size and resource caps are listed under [Constraints](#constraints) (`internal/pdf/compress/limits.go`).

## Where it runs

**Only in the visitor’s browser.** You host static files (`compress.wasm`, `wasm_exec.js`, a few lines of JS). The PDF is read in the tab, compressed in the tab, and downloaded from the tab. It does not go to gopdfsuit’s API and it does not go to your backend unless you add that yourself.

A closed-source SaaS (an iLovePDF-style site) can ship this as a static asset on its own origin. Each user compresses on their own machine. Your server’s job is CDN/hosting, not PDF processing.

`goCompressPDF` is a page global after WASM loads. Any script on **that same origin** can call it. That is still client-side.

## Run

```bash
make wasm-compress                 # from repo root, once (writes frontend/public/)
cp frontend/public/compress.wasm frontend/public/wasm_exec.js sampledata/compress-js/
node sampledata/compress-js/run.mjs
```

The local `compress.wasm` / `wasm_exec.js` copies are gitignored work
products, never committed (row 5.6). Re-copy after rebuilding.

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

## Constraints

Defined in `internal/pdf/compress/limits.go`. They apply to **this WASM path, `gopdflib.CompressPDF`, and `POST /api/v1/compress`**. A hostile file should fail closed (error, or skip that stream) instead of growing without bound. They bound **this user’s tab** (or one API request); they do not make compress instant.

### Size and resource caps

| Constraint | Limit | What it stops |
|------------|--------|----------------|
| Input size | **32 MiB** | Huge upload / `Uint8Array`. Checked in JS, WASM (`byteLength` + `CopyBytesToGo`), `CompressPDF`, and HTTP `LimitReader`. |
| Flate inflate | **48 MiB per stream** | Zip bomb (tiny Flate → hundreds of MB). Same cap in `compress.decompressFlate` and `merge.decompressFlate` (object streams). |
| Image raster | **16 megapixels**, longest edge **8192** | `/Width 100000 /Height 100000` or a huge JPEG. JPEG size is read with `DecodeConfig` *before* full decode. |
| `max_image_dim` override | **4096** | HTTP `max_image_dim=999999999` cannot disable downscale. |
| JPEG quality | **1–100** | Values over 100 are clamped. |
| PDF objects | **50_000** count **and** object number | A single `50001 0 obj` used to make xref writing loop `1..N`. Object streams with `/N` above this are ignored. |

The 32 MiB ceiling is the main size number. Inside it, a awkward file can still freeze the tab for a while.

### Compression tiers (quality presets, not DoS caps)

| Level | JPEG quality | Max image edge |
|-------|----------------|----------------|
| Light (`1`) | 92 | 1920 |
| Medium (`2`, default) | 75 | 1275 |
| Heavy (`3`) | 50 | 612 |

### Fail-closed rules

- Encrypted PDFs (`/Encrypt`) are **rejected**, not decrypted.
- Non-PDF / empty input is **rejected**.
- WASM accepts a **`Uint8Array` only** (a fake `{ byteLength: huge }` copies 0 bytes and fails).
- WASM **panics** become `{ error: "compression failed" }` instead of killing the tab.
- JPEG2000, JBIG2, SMask, `/DecodeParms`, odd bit depths are **skipped**, not decoded.

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

No Ghostscript (or other AGPL) is used. That is the license that would have forced a network SaaS to publish source. This path does not include it.

The compressor is this repo’s MIT engine (`internal/pdf/compress`), the same code as `gopdflib.CompressPDF`. Closed-source products can use MIT and BSD-3-Clause. Neither is copyleft. You do **not** have to open-source the host app.

| Piece | License | Closed-source SaaS |
|-------|---------|-------------------|
| Engine, JS wrapper, this sample | MIT (this repository) | Allowed. Keep the MIT copyright notice. |
| `wasm_exec.js` | BSD-3-Clause, Copyright The Go Authors | Allowed. Keep the file’s Go copyright header. Do not strip it. |
| `compress.wasm` | Our MIT package plus the Go WASM runtime (BSD-3-Clause) | Allowed. That runtime is how `GOOS=js GOARCH=wasm` always works. |

BSD-3-Clause is a **notice** license, not a “share your product” license. In practice:

1. Leave the Go header at the top of `wasm_exec.js`.
2. In LICENSE, NOTICE, or an About page, note that the product includes Go’s WASM runtime under BSD-3-Clause.
3. Do not advertise the product as endorsed by Google or the Go Authors (the third BSD clause).

You do not need a Ghostscript license, a GPL/AGPL grant, or to publish your SaaS source because of this WASM. This is not legal advice; it is the usual reading of MIT + BSD-3.
