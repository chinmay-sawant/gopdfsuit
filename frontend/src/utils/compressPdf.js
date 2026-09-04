// Re-export shim: compress now lives in ./wasm/compress.js (levels in
// ./wasm/levels.js, worker in ./wasm/compressWorker.js). New code should
// import from there directly.
export {
  MAX_COMPRESS_BYTES,
  compressViaServer,
  compressPDF,
  compressPDFSmart,
} from './wasm/compress.js'
