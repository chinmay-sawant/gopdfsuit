// Compliant generate preset: fonts first, then a pdfaCompliant template.
// veraPDF validation itself stays server-side; this module produces the
// bytes that a server check (or POST /api/v1/generate/template-pdf with the
// same JSON) can verify.

import { ensureGopdfsuitWasm } from './core.js'
import { ensurePDFAFonts } from './fonts.js'

function callWasmPdf(fnName, args) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') {
    const err = new Error(`${fnName} is not in the shipped WASM bundle yet`)
    err.fallbackAvailable = true
    err.missingEngine = true
    throw err
  }
  const result = fn(...args)
  if (result instanceof Uint8Array) return result
  const message = result && typeof result === 'object' ? result.error || result.message : undefined
  throw new Error(message || `${fnName} failed`)
}

function withPDFACompliant(template) {
  if (typeof template === 'string') {
    const parsed = JSON.parse(template)
    parsed.config = { ...(parsed.config || {}), pdfaCompliant: true }
    return parsed
  }
  return { ...template, config: { ...((template && template.config) || {}), pdfaCompliant: true } }
}

/**
 * Generate a PDF/A-compliant PDF in the browser. Ensures the 12 Liberation
 * faces are registered (fetching from /fonts/ on first use), forces
 * pdfaCompliant:true on a template copy, and generates. Throws listing
 * still-missing faces when the font fetch fails.
 */
export async function generateCompliantPDF(template) {
  await ensureGopdfsuitWasm()
  const fonts = await ensurePDFAFonts()
  if (fonts.missing.length > 0) {
    throw new Error(`PDFA fonts unavailable: ${fonts.missing.join(', ')}`)
  }
  return callWasmPdf('goGeneratePDF', [withPDFACompliant(template)])
}
