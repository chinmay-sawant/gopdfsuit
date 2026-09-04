// Pins the Go WASM {code,message,error} envelope shape owned by
// src/utils/wasm/envelope.js (Phase 3.2). Run with `npm test`
// (node --test tests/).
import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

import {
  callWasm,
  callWasmObject,
  missingEngineError,
  normalizeWasmResult,
  wasmErrorMessage,
} from '../src/utils/wasm/envelope.js'

describe('wasm envelope', () => {
  it('passes Uint8Array results through untouched', () => {
    const bytes = new Uint8Array([1, 2, 3])
    globalThis.goTestBytes = () => bytes
    assert.equal(callWasm('goTestBytes', []), bytes)
    delete globalThis.goTestBytes
  })

  it('passes arrays of Uint8Array only with allowArray', () => {
    const parts = [new Uint8Array([1]), new Uint8Array([2])]
    globalThis.goTestParts = () => parts
    assert.deepEqual(callWasm('goTestParts', [], { allowArray: true }), parts)
    assert.throws(() => callWasm('goTestParts', []), /goTestParts failed/)
    delete globalThis.goTestParts
  })

  it('maps {code,message,error} failures to Error(message)', () => {
    globalThis.goTestFail = () => ({ code: 'INVALID_INPUT', message: 'empty PDF', error: 'empty PDF' })
    assert.throws(() => callWasm('goTestFail', []), /empty PDF/)
    delete globalThis.goTestFail
  })

  it('prefers message over the legacy error alias', () => {
    assert.equal(
      wasmErrorMessage({ code: 'X', message: 'new', error: 'old' }, 'fallback'),
      'new',
    )
    assert.equal(wasmErrorMessage({ code: 'X', error: 'old' }, 'fallback'), 'old')
    assert.equal(wasmErrorMessage(null, 'fallback'), 'fallback')
  })

  it('marks missing-engine errors as consentable', () => {
    const err = missingEngineError('goNope')
    assert.match(err.message, /goNope/)
    assert.equal(err.fallbackAvailable, true)
    assert.equal(err.missingEngine, true)
    assert.throws(() => callWasm('goDefinitelyMissing', []), /goDefinitelyMissing/)
    try {
      callWasm('goDefinitelyMissing', [])
    } catch (err2) {
      assert.equal(err2.fallbackAvailable, true)
    }
  })

  it('callWasmObject passes status objects through but throws envelopes', () => {
    globalThis.goTestStatus = () => ({ registered: ['Helvetica'], missing: [] })
    assert.deepEqual(callWasmObject('goTestStatus', []), { registered: ['Helvetica'], missing: [] })
    globalThis.goTestStatus = () => ({ code: 'X', message: 'bad', error: 'bad' })
    assert.throws(() => callWasmObject('goTestStatus', []), /bad/)
    delete globalThis.goTestStatus
  })

  it('normalizeWasmResult rejects non-bytes without an envelope', () => {
    assert.throws(() => normalizeWasmResult('goX', 42), /goX failed/)
    assert.throws(() => normalizeWasmResult('goX', 'oops'), /goX failed/)
  })
})
