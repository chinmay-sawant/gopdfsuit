// Template generate via gopdfsuit.wasm. Accepts a template object or JSON
// string; the Go shim stringifies objects itself, so callers pass through.
// Large binary assets stay base64 imagedata/fontData strings, matching the
// server template contract. For the compliant variant (fonts ensured first)
// see ./compliance.js.

import { ensureGopdfsuitWasm, callWasm } from './core.js'
import { makeAuthenticatedRequest } from '../apiConfig.js'
import { ensurePDFAFonts } from './fonts.js'
import { smartLocal } from './transports.js'

function callWasmPdf(fnName, args) {
  return callWasm(fnName, args)
}

export async function generatePDFViaWasm(template) {
  await ensureGopdfsuitWasm()
  if (isPDFACompliant(template)) {
    // Compliant output needs embedded subsets: ensure the Liberation faces
    // are registered before generating (fetched once, then Cache API).
    await ensurePDFAFonts()
  }
  return callWasmPdf('goGeneratePDF', [template])
}

function isPDFACompliant(template) {
  try {
    const obj = typeof template === 'string' ? JSON.parse(template) : template
    return Boolean(obj && obj.config && obj.config.pdfaCompliant)
  } catch {
    return false
  }
}

export async function generateViaServer(template, getAuthHeaders) {
  const body = typeof template === 'string' ? template : JSON.stringify(template)
  const response = await makeAuthenticatedRequest(
    '/api/v1/generate/template-pdf',
    { method: 'POST', body, headers: { 'Content-Type': 'application/json' } },
    getAuthHeaders,
  )
  const blob = await response.blob()
  const out = new Uint8Array(await blob.arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return out
}

export const generatePDFSmart = (template, transport = {}) =>
  smartLocal(
    () => generatePDFViaWasm(template),
    () => generateViaServer(template, transport.getAuthHeaders),
    { ...transport, getAuthHeaders: transport.getAuthHeaders },
  )
