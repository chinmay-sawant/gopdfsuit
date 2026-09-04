// Re-export shim: levels now live in ./wasm/levels.js. New code should
// import from there directly.
export {
  MAX_COMPRESS_BYTES,
  COMPRESS_LEVELS,
  DEFAULT_COMPRESS_LEVEL,
  levelByValue,
  toServerLevel,
  toWasmLevel,
  assertCompressSize,
  COMPRESS_TRANSPORT,
  shouldUseServerCompress,
  default,
} from './wasm/levels.js'
