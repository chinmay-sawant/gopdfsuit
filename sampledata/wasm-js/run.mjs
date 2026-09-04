// Full-port WASM smoke: generate, merge, split, fill, redact via gopdfsuit.wasm.
// Requires gopdfsuit.wasm + wasm_exec.js next to this file or in frontend/public/.
// Build them with: make wasm
// Usage: node sampledata/wasm-js/run.mjs
//
// Mirrors sampledata/compress-js/run.mjs. No HTTP, no build execution here:
// this file only wires the five bindings; the Go entrypoint lands in
// cmd/wasm (plans/wasm/01-full-wasm-port.md Phase 2.1).

import { readFile, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { applyRedactionsAdvanced, fillPDF, findTextOccurrences, generatePDF, htmlToImage, htmlToPDF, mergePDFs, splitPDF } from './gopdfsuit.js'

const dir = dirname(fileURLToPath(import.meta.url))

const template = {
  config: { page: 'A4', pageAlignment: 1, pdfTitle: 'WASM smoke' },
  title: { props: 'Helvetica:18:100:center:0:0:0:0', text: 'WASM SMOKE' },
  elements: [
    {
      type: 'table',
      table: {
        maxcolumns: 2,
        columnwidths: [1, 2],
        rows: [
          { row: [{ props: 'Helvetica:10:100:left:1:1:1:1', text: 'Company:' }, { props: 'Helvetica:10:000:left:1:1:1:1', text: 'WasmSmoke Inc.' }] },
          { row: [{ props: 'Helvetica:10:100:left:1:1:1:1', text: 'Note:' }, { props: 'Helvetica:10:000:left:1:1:1:1', text: 'RedactMe marker line' }] },
        ],
      },
    },
  ],
  footer: { font: 'Helvetica:8:000:center', text: 'wasm-js smoke' },
}

try {
  // 1. Generate from a template object.
  const generated = await generatePDF(template)
  await writeFile(join(dir, 'generated.pdf'), generated)
  console.log(`generate  generated.pdf  ${generated.length} bytes`)

  // 2. Merge two fixtures.
  const [a, b] = await Promise.all([
    readFile(join(dir, '..', 'merge', 'em-16.pdf')),
    readFile(join(dir, '..', 'merge', 'em-19.pdf')),
  ])
  const merged = await mergePDFs([a, b])
  await writeFile(join(dir, 'merged.pdf'), merged)
  console.log(`merge     merged.pdf  ${merged.length} bytes from em-16.pdf + em-19.pdf`)

  // 3. Split the engine-generated PDF into single pages (JS array of
  // Uint8Array, no Go-side zip). Note: merge output is not splittable -
  // the merge parser cannot round-trip MergePDFs bytes (pre-existing).
  const parts = await splitPDF(generated, { maxPerFile: 1 })
  for (let i = 0; i < parts.length; i += 1) {
    await writeFile(join(dir, `split_part_${i + 1}.pdf`), parts[i])
  }
  console.log(`split     ${parts.length} part(s) written as split_part_N.pdf`)

  // 4. Fill the hospital AcroForm fixture with its XFDF fixture.
  const [formPdf, xfdf] = await Promise.all([
    readFile(join(dir, '..', 'filler', 'us_hospital_encounter_acroform.pdf')),
    readFile(join(dir, '..', 'filler', 'us_hospital_encounter_data.xfdf')),
  ])
  const filled = await fillPDF(formPdf, xfdf)
  await writeFile(join(dir, 'filled.pdf'), filled)
  console.log(`fill      filled.pdf  ${filled.length} bytes`)

  // 5. Redact via the text path (no OCR in WASM).
  const hits = await findTextOccurrences(generated, 'RedactMe')
  console.log(`redact    search hits: ${JSON.stringify(hits)?.slice(0, 200)}`)
  const { pdf: redacted } = await applyRedactionsAdvanced(generated, { textSearch: [{ text: 'RedactMe' }], mode: 'secure_required' })
  await writeFile(join(dir, 'redacted.pdf'), redacted)
  console.log(`redact    redacted.pdf  ${redacted.length} bytes`)

  // 6. HTML to PDF/Image from inline strings (URLs stay server-side).
  const demoHtml = '<html><body><h1>WASM HTML invoice</h1><p>Rendered in-browser, no upload.</p></body></html>'
  const htmlPdf = await htmlToPDF(demoHtml, { page_size: 'A4', orientation: 'Portrait' })
  await writeFile(join(dir, 'html.pdf'), htmlPdf)
  console.log(`html-pdf  html.pdf  ${htmlPdf.length} bytes`)
  const htmlPng = await htmlToImage(demoHtml, { format: 'png', width: 800, height: 600 })
  await writeFile(join(dir, 'html.png'), htmlPng)
  console.log(`html-img  html.png  ${htmlPng.length} bytes`)
} catch (err) {
  const msg = err?.message ?? String(err)
  console.error(msg)
  if (!msg.includes('make wasm')) {
    console.error('If gopdfsuit.wasm is missing, run `make wasm` first.')
  }
  process.exit(1)
}
