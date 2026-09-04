# PDF compress — Go library

`gopdflib.CompressPDF` sample. No Ghostscript, no WASM, no HTTP.

```bash
cd sampledata/compress && go run .
```

Reads `report.pdf` and writes:

| File | Level |
|------|--------|
| `report_level_1.pdf` | Light (JPEG 92, max edge 1920) |
| `report_level_2.pdf` | Medium (JPEG 75, max edge 1275) |
| `report_level_3.pdf` | Heavy (JPEG 50, max edge 612) |

In-browser WASM lives in [`sampledata/compress-js`](../compress-js).
