// Compliant-generate smoke: register the 12 Liberation faces, then generate
// sampledata/financialreport/financial_report.json with pdfaCompliant:true
// through gopdfsuit.wasm. Usage: node sampledata/wasm-js/run_compliant.mjs
// Fonts come from frontend/public/fonts/ (vendored, OFL NOTICE included).

import { readFile, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { ensurePDFAFonts, generatePDF, registerFont } from './gopdfsuit.js'

const dir = dirname(fileURLToPath(import.meta.url))
const repoRoot = join(dir, '..', '..')
const fontsDir = join(repoRoot, 'frontend', 'public', 'fonts')

const MANIFEST = [
  ['Helvetica', 'LiberationSans-Regular.ttf'],
  ['Helvetica-Bold', 'LiberationSans-Bold.ttf'],
  ['Helvetica-Oblique', 'LiberationSans-Italic.ttf'],
  ['Helvetica-BoldOblique', 'LiberationSans-BoldItalic.ttf'],
  ['Times-Roman', 'LiberationSerif-Regular.ttf'],
  ['Times-Bold', 'LiberationSerif-Bold.ttf'],
  ['Times-Italic', 'LiberationSerif-Italic.ttf'],
  ['Times-BoldItalic', 'LiberationSerif-BoldItalic.ttf'],
  ['Courier', 'LiberationMono-Regular.ttf'],
  ['Courier-Bold', 'LiberationMono-Bold.ttf'],
  ['Courier-Oblique', 'LiberationMono-Italic.ttf'],
  ['Courier-BoldOblique', 'LiberationMono-BoldItalic.ttf'],
]

try {
  for (const [name, file] of MANIFEST) {
    const bytes = await readFile(join(fontsDir, file))
    await registerFont(name, bytes)
  }
  const status = await ensurePDFAFonts()
  console.log(`fonts registered=${status.registered.length} missing=${status.missing.length}`)
  if (status.missing.length > 0) throw new Error(`missing faces: ${status.missing.join(', ')}`)

  const raw = await readFile(join(repoRoot, 'sampledata', 'financialreport', 'financial_report.json'), 'utf8')
  const template = JSON.parse(raw)
  // pdfaCompliant triggers subset embedding; pdfTitle feeds XMP dc:title,
  // which PDF/UA-2 requires (same contract as the server endpoint).
  template.config = { ...(template.config || {}), pdfaCompliant: true, pdfTitle: 'Financial Report Q4 2025' }
  const out = await generatePDF(template)
  await writeFile(join(dir, 'compliant.pdf'), out)
  console.log(`compliant  compliant.pdf  ${out.length} bytes`)

  const head = Buffer.from(out.slice(0, 5)).toString('ascii')
  const tail = Buffer.from(out.slice(-6)).toString('ascii')
  if (!head.startsWith('%PDF-')) throw new Error(`bad header: ${head}`)
  if (!tail.includes('%%EOF')) throw new Error('missing %%EOF marker')
  const text = Buffer.from(out).toString('latin1')
  const fontFiles = (text.match(/\/FontFile2? /g) || []).length
  console.log(`header=${head} eof-ok font-subsets=${fontFiles}`)
  if (fontFiles === 0) throw new Error('no embedded font subsets found')
} catch (err) {
  console.error(err?.message ?? String(err))
  process.exit(1)
}
