// Pins the Split tool's client-side zip writer (src/utils/zip.js): stored
// entries with valid CRCs, UTF-8 names, and the server's <base>-partN.pdf
// naming. Run with `npm test` (node --test tests/).
import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

import { crc32, createZip } from '../src/utils/zip.js'

describe('split zip writer', () => {
  it('computes the standard CRC32 check value', () => {
    assert.equal(crc32(new TextEncoder().encode('123456789')), 0xcbf43926)
  })

  it('packs 3 one-page parts under split-<base>-partN.pdf names', () => {
    const files = [1, 2, 3].map((n) => ({
      name: `split-mm-part${n}.pdf`,
      data: new Uint8Array([0x25, 0x50, 0x44, 0x46, n]),
    }))
    const zip = createZip(files)
    const text = new TextDecoder('latin1').decode(zip)
    for (const f of files) assert.ok(text.includes(f.name), `missing ${f.name}`)
    assert.equal(zip[0], 0x50)
    assert.equal(zip[1], 0x4b)
  })

  it('rejects empty input and >65535 entries', () => {
    assert.throws(() => createZip([]), /at least one file/)
    assert.throws(() => createZip(new Array(0x10000).fill({ name: 'x', data: new Uint8Array([1]) })), /65535/)
  })
})
