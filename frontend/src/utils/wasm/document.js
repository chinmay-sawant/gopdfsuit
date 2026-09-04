// Merge / Split / Fill via gopdfsuit.wasm, with explicit-consent server
// fallbacks. Split returns a JS array of Uint8Array (multi-file download is
// assembled by usePdfOperation.runLocalMulti); zipping stays in JS so
// archive/zip never enters the WASM closure.

import { ensureGopdfsuitWasm, asUint8Array } from './core.js'
import { makeAuthenticatedRequest } from '../apiConfig.js'
import { smartLocal } from './transports.js'

function missingEngineError(fnName) {
  const err = new Error(
    `${fnName} is not in the shipped WASM bundle yet (needs plans/wasm/01-full-wasm-port.md Fill/Merge/Split bindings)`,
  )
  err.fallbackAvailable = true
  err.missingEngine = true
  return err
}

function callWasmArray(fnName, args) {
  const fn = globalThis[fnName]
  if (typeof fn !== 'function') throw missingEngineError(fnName)
  const result = fn(...args)
  if (result instanceof Uint8Array) return result
  if (Array.isArray(result) && result.every((entry) => entry instanceof Uint8Array)) return result
  const message = result && typeof result === 'object' ? result.error || result.message : undefined
  throw new Error(message || `${fnName} failed`)
}

export async function mergePDFViaWasm(files) {
  await ensureGopdfsuitWasm()
  const parts = await Promise.all(
    files.map(async (file) => asUint8Array(await file.arrayBuffer())),
  )
  return callWasmArray('goMergePDF', [parts])
}

export async function splitPDFViaWasm(bytes, { pages = '', maxPerFile = '' } = {}) {
  await ensureGopdfsuitWasm()
  return callWasmArray('goSplitPDF', [asUint8Array(bytes), String(pages || ''), String(maxPerFile || '')])
}

export async function fillPDFViaWasm(pdfBytes, xfdfBytes) {
  await ensureGopdfsuitWasm()
  return callWasmArray('goFillPDF', [asUint8Array(pdfBytes), asUint8Array(xfdfBytes)])
}

export async function mergeViaServer(files, getAuthHeaders) {
  const formData = new FormData()
  files.forEach((file) => formData.append('pdf', file))
  const response = await makeAuthenticatedRequest('/api/v1/merge', { method: 'POST', body: formData }, getAuthHeaders)
  const out = new Uint8Array(await (await response.blob()).arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return out
}

export async function splitViaServer(bytes, { pages = '', maxPerFile = '' } = {}, getAuthHeaders) {
  const formData = new FormData()
  formData.append('pdf', new Blob([asUint8Array(bytes)], { type: 'application/pdf' }), 'document.pdf')
  if (pages) formData.append('pages', pages)
  if (maxPerFile) formData.append('max_per_file', maxPerFile)
  const response = await makeAuthenticatedRequest('/api/v1/split', { method: 'POST', body: formData }, getAuthHeaders)
  const out = new Uint8Array(await (await response.blob()).arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return [out]
}

export async function fillViaServer(pdfBytes, xfdfBytes, pdfName, getAuthHeaders) {
  const formData = new FormData()
  formData.append('pdf', new Blob([asUint8Array(pdfBytes)], { type: 'application/pdf' }), pdfName || 'form.pdf')
  formData.append('xfdf', new Blob([asUint8Array(xfdfBytes)], { type: 'application/xml' }), 'data.xfdf')
  const response = await makeAuthenticatedRequest('/api/v1/fill', { method: 'POST', body: formData }, getAuthHeaders)
  const out = new Uint8Array(await (await response.blob()).arrayBuffer())
  if (out.byteLength === 0) throw new Error('Received empty document')
  return out
}

export const mergePDFSmart = (files, opts, transport = {}) =>
  smartLocal(() => mergePDFViaWasm(files), () => mergeViaServer(files, transport.getAuthHeaders), {
    ...transport,
    getAuthHeaders: transport.getAuthHeaders,
  })

export const splitPDFSmart = (bytes, splitOpts, transport = {}) =>
  smartLocal(
    () => splitPDFViaWasm(bytes, splitOpts),
    () => splitViaServer(bytes, splitOpts, transport.getAuthHeaders),
    { ...transport, getAuthHeaders: transport.getAuthHeaders },
  )

export const fillPDFSmart = (pdfBytes, xfdfBytes, pdfName, transport = {}) =>
  smartLocal(
    () => fillPDFViaWasm(pdfBytes, xfdfBytes),
    () => fillViaServer(pdfBytes, xfdfBytes, pdfName, transport.getAuthHeaders),
    { ...transport, getAuthHeaders: transport.getAuthHeaders },
  )
