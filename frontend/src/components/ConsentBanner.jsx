/**
 * ConsentBanner owns the shared explicit-consent banner: browser-local WASM
 * failed without uploading anything, and the user must click to upload to
 * the server. Rendered from hook-owned state (usePdfOperation consentOffer);
 * nothing uploads silently.
 */
const ConsentBanner = ({ offer, onConsent, onDismiss, isLoading = false, actionLabel = 'Upload to server' }) => {
  if (!offer) return null
  return (
    <div style={{ padding: '1rem', background: 'rgba(255, 193, 7, 0.1)', border: '1px solid #ffc107', borderRadius: '8px', marginBottom: '1rem', color: 'hsl(var(--foreground))' }}>
      <div style={{ marginBottom: '0.75rem' }}>
        Browser processing is not available in this build{offer.message ? `: ${offer.message}` : '.'} The file was not uploaded.
        Upload it to the server instead?
      </div>
      <div style={{ display: 'flex', gap: '0.75rem' }}>
        <button onClick={onConsent} disabled={isLoading} className="btn-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
          {actionLabel}
        </button>
        <button onClick={onDismiss} disabled={isLoading} className="btn-outline-glow" style={{ padding: '0.5rem 1rem', fontSize: '0.9rem' }}>
          Stay local
        </button>
      </div>
    </div>
  )
}

export default ConsentBanner
