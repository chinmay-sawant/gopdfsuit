# ADR Stub: Compress Server-Fallback Default

> **Status:** question open - owner: chinmay-sawant
> **Context:** round2 E5 (`plans/reviews/architecture-review-2026-09-04-round2.md` Grill Decisions)

## Current behavior (shipped)

- `VITE_COMPRESS_TRANSPORT` (`frontend/src/utils/compressLevels.js`) selects the
  transport; default `wasm` keeps the file on-device.
- When browser (WASM) compression fails, the UI shows an explicit fallback
  notice and only uploads to `/api/v1/compress` after a consent click
  (`frontend/src/pages/Compress.jsx`, `compressPDFSmart` with
  `allowServerFallback: false` by default).
- Server transport (`VITE_COMPRESS_TRANSPORT=server`) is surfaced in the UI
  with an upload notice.

## Open product question

Should the server fallback stay consent-gated, or become automatic (silent
upload on WASM failure) for some deployment default?

Options:

1. Keep consent-gated everywhere (current).
2. Auto-fallback on self-hosted builds, consent-gated on the public demo.
3. Per-user preference persisted in localStorage.

Decision and rationale to be recorded here by chinmay-sawant.
