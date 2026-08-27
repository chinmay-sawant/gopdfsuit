// Compress report.pdf at levels 1–3 via Go WASM (same engine as sampledata/compress).
// Requires compress.wasm + wasm_exec.js next to this file or in frontend/public/.
// Build them with: make wasm-compress
// Usage: node sampledata/compress-js/run.mjs

import { readFile, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { compressPDF } from './compress.js'

const dir = dirname(fileURLToPath(import.meta.url))
const srcName = 'report.pdf'
const srcPath = join(dir, srcName)

const levels = [
  { n: 1, name: 'Light', file: 'report_js_level_1.pdf' },
  { n: 2, name: 'Medium', file: 'report_js_level_2.pdf' },
  { n: 3, name: 'Heavy', file: 'report_js_level_3.pdf' },
]

try {
  const src = await readFile(srcPath)
  console.log(`source  ${srcName}  ${src.length} bytes`)

  for (const lv of levels) {
    const out = await compressPDF(src, { level: lv.n })
    await writeFile(join(dir, lv.file), out)
    const pct = 100 - (out.length * 100) / src.length
    console.log(
      `level ${lv.n} ${lv.name}  ${lv.file}  ${out.length} bytes  (${pct.toFixed(1)}% smaller)`,
    )
  }
} catch (err) {
  const msg = err?.message ?? String(err)
  console.error(msg)
  if (!msg.includes('make wasm-compress')) {
    console.error('If compress.wasm is missing, run `make wasm-compress` first.')
  }
  process.exit(1)
}
