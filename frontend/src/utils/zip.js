// Minimal ZIP writer (stored entries, no compression) for the Split tool's
// "Download all (.zip)" action. Hand-rolled so the frontend gains no new
// dependency: entries are stored, CRC32 is computed inline, and the output
// is a plain Uint8Array ready for Blob download. Symmetric with the
// server's splits.zip (archive/zip) naming: <base>-partN.pdf.

const CRC_TABLE = (() => {
  const table = new Uint32Array(256)
  for (let n = 0; n < 256; n++) {
    let c = n
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
    table[n] = c >>> 0
  }
  return table
})()

export function crc32(data) {
  let crc = 0xffffffff
  for (let i = 0; i < data.length; i++) crc = CRC_TABLE[(crc ^ data[i]) & 0xff] ^ (crc >>> 8)
  return (crc ^ 0xffffffff) >>> 0
}

const encoder = new TextEncoder()

function dosTime(date = new Date()) {
  const time = ((date.getHours() & 0x1f) << 11) | ((date.getMinutes() & 0x3f) << 5) | ((date.getSeconds() >> 1) & 0x1f)
  const day = ((date.getFullYear() - 1980) << 9) | ((date.getMonth() + 1) << 5) | (date.getDate() & 0x1f)
  return { time, day }
}

// files: [{ name: string, data: Uint8Array }]. Throws on empty input or on
// more than 65535 entries (16-bit end record, same ceiling browsers hit).
export function createZip(files) {
  if (!Array.isArray(files) || files.length === 0) throw new Error('createZip needs at least one file')
  if (files.length > 0xffff) throw new Error('createZip supports at most 65535 files')
  const { time, day } = dosTime()
  const chunks = []
  const central = []
  let offset = 0
  const push = (bytes) => {
    chunks.push(bytes)
    offset += bytes.length
  }

  files.forEach(({ name, data }) => {
    const nameBytes = encoder.encode(name)
    const bytes = data instanceof Uint8Array ? data : new Uint8Array(data)
    const crc = crc32(bytes)
    const headerOffset = offset
    const local = new DataView(new ArrayBuffer(30))
    local.setUint32(0, 0x04034b50, true)
    local.setUint16(4, 20, true)
    local.setUint16(6, 0x0800, true) // UTF-8 filenames
    local.setUint16(8, 0, true) // stored
    local.setUint16(10, time, true)
    local.setUint16(12, day, true)
    local.setUint32(14, crc, true)
    local.setUint32(18, bytes.length, true)
    local.setUint32(22, bytes.length, true)
    local.setUint16(26, nameBytes.length, true)
    local.setUint16(28, 0, true)
    push(new Uint8Array(local.buffer))
    push(nameBytes)
    push(bytes)

    const entry = new DataView(new ArrayBuffer(46))
    entry.setUint32(0, 0x02014b50, true)
    entry.setUint16(4, 20, true)
    entry.setUint16(6, 20, true)
    entry.setUint16(8, 0x0800, true)
    entry.setUint16(10, 0, true)
    entry.setUint16(12, time, true)
    entry.setUint16(14, day, true)
    entry.setUint32(16, crc, true)
    entry.setUint32(20, bytes.length, true)
    entry.setUint32(24, bytes.length, true)
    entry.setUint16(28, nameBytes.length, true)
    entry.setUint16(30, 0, true)
    entry.setUint16(32, 0, true)
    entry.setUint16(34, 0, true)
    entry.setUint16(36, 0, true)
    entry.setUint32(38, 0, true)
    entry.setUint32(42, headerOffset, true)
    central.push(new Uint8Array(entry.buffer), nameBytes)
  })

  const centralOffset = offset
  let centralSize = 0
  central.forEach((part) => {
    chunks.push(part)
    centralSize += part.length
  })
  offset += centralSize

  const end = new DataView(new ArrayBuffer(22))
  end.setUint32(0, 0x06054b50, true)
  end.setUint16(4, 0, true)
  end.setUint16(6, 0, true)
  end.setUint16(8, files.length, true)
  end.setUint16(10, files.length, true)
  end.setUint32(12, centralSize, true)
  end.setUint32(16, centralOffset, true)
  end.setUint16(20, 0, true)
  chunks.push(new Uint8Array(end.buffer))
  offset += 22

  const out = new Uint8Array(offset)
  let at = 0
  chunks.forEach((part) => {
    out.set(part, at)
    at += part.length
  })
  return out
}
