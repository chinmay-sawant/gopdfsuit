// Compress PDF bytes off the main thread via Go WASM (classic worker).
// Instantiated from compressPdf.js with new URL('./compressWorker.js',
// import.meta.url); Vite bundles this file and rewrites the URL. Main thread
// fallback stays in compressPdf.js when Worker is missing or fails.
//
// Protocol (all messages are plain objects):
//   main -> worker { type: 'init', wasmExecUrl, wasmUrl }
//   worker -> main { type: 'ready' } | { type: 'init-error', error }
//   main -> worker { type: 'compress', id, level, bytes } (bytes transferred)
//   worker -> main { type: 'result', id, ok, bytes?, error? } (bytes transferred)

let goRun = null
let initError = null

async function init(wasmExecUrl, wasmUrl) {
  try {
    self.importScripts(wasmExecUrl)
    const Go = self.Go
    if (typeof Go !== 'function') {
      throw new Error('wasm_exec.js did not provide global Go')
    }
    const go = new Go()
    const response = await fetch(wasmUrl)
    if (!response.ok) {
      throw new Error(`compress.wasm fetch failed: ${response.status}`)
    }
    let instance
    if (typeof WebAssembly.instantiateStreaming === 'function') {
      try {
        const streamed = await WebAssembly.instantiateStreaming(
          response.clone(),
          go.importObject,
        )
        instance = streamed.instance
      } catch {
        const buffer = await response.arrayBuffer()
        const built = await WebAssembly.instantiate(buffer, go.importObject)
        instance = built.instance
      }
    } else {
      const buffer = await response.arrayBuffer()
      const built = await WebAssembly.instantiate(buffer, go.importObject)
      instance = built.instance
    }
    go.run(instance)
    if (typeof self.goCompressPDF !== 'function') {
      throw new Error('goCompressPDF not registered after go.run')
    }
    goRun = true
    self.postMessage({ type: 'ready' })
  } catch (err) {
    initError = err instanceof Error ? err.message : String(err)
    self.postMessage({ type: 'init-error', error: initError })
  }
}

self.onmessage = (event) => {
  const msg = event.data || {}
  if (msg.type === 'init') {
    init(msg.wasmExecUrl, msg.wasmUrl)
    return
  }
  if (msg.type !== 'compress') return
  const { id, level, bytes } = msg
  try {
    if (initError) throw new Error(initError)
    if (!goRun || typeof self.goCompressPDF !== 'function') {
      throw new Error('WASM not ready in worker')
    }
    const input = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
    const result = self.goCompressPDF(input, level)
    if (result instanceof Uint8Array) {
      self.postMessage({ type: 'result', id, ok: true, bytes: result }, [result.buffer])
      return
    }
    const message = result && typeof result === 'object' ? result.error : undefined
    self.postMessage({ type: 'result', id, ok: false, error: message || 'PDF compression failed' })
  } catch (err) {
    self.postMessage({ type: 'result', id, ok: false, error: err instanceof Error ? err.message : String(err) })
  }
}
