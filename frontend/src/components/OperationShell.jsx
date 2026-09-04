import { Download } from 'lucide-react'

/**
 * OperationShell owns the shared result panel for PDF operations:
 * preview (PDF iframe or image), optional stats row, download action,
 * and the empty placeholder shown before the first run.
 */
const OperationShell = ({
  resultUrl,
  title,
  icon,
  emptyTitle,
  emptySubtitle,
  onDownload,
  downloadLabel = 'Download',
  height = 550,
  kind = 'pdf',
  imageAlt = 'Generated output',
  stats = null,
  isLoading = false,
  loadingLabel = '',
}) => (
  <div className="glass-card" style={{ padding: '2rem' }}>
    <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.2rem', fontWeight: '700' }}>
      <div className="feature-icon-box purple" style={{ width: '40px', height: '40px', marginBottom: 0 }}>{icon}</div>{title}
    </h3>
    {resultUrl ? (
      <div>
        {stats}
        {kind === 'image' ? (
          <div style={{ border: '1px solid rgba(255,255,255,0.15)', borderRadius: '8px', padding: '1rem', textAlign: 'center', background: 'rgba(255,255,255,0.02)', marginBottom: '1rem' }}>
            <img src={resultUrl} alt={imageAlt} style={{ maxWidth: '100%', maxHeight: `${height}px`, borderRadius: '6px', boxShadow: '0 4px 8px rgba(0,0,0,0.3)' }} />
          </div>
        ) : (
          <iframe src={resultUrl} style={{ width: '100%', height: `${height}px`, border: '1px solid rgba(255,255,255,0.15)', borderRadius: '8px', overflow: 'hidden' }} title={title} />
        )}
        <button
          onClick={() => { if (onDownload) onDownload() }}
          disabled={isLoading}
          className="btn-glow"
          style={{ width: '100%', marginTop: '1rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '0.75rem 1.5rem' }}
        >
          <Download size={16} />{isLoading && loadingLabel ? loadingLabel : downloadLabel}
        </button>
      </div>
    ) : (
      <div style={{ height: `${height}px`, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(255,255,255,0.02)', borderRadius: '8px', border: '2px dashed rgba(255,255,255,0.1)', color: 'hsl(var(--muted-foreground))', textAlign: 'center' }}>
        <div>
          <div className="feature-icon-box purple" style={{ width: '64px', height: '64px', margin: '0 auto 1rem', opacity: 0.5 }}>{icon}</div>
          <p style={{ marginBottom: '0.5rem', fontSize: '1.1rem', fontWeight: '600' }}>{emptyTitle}</p>
          <p style={{ fontSize: '0.9rem', opacity: 0.7, marginBottom: 0 }}>{emptySubtitle}</p>
        </div>
      </div>
    )}
  </div>
)

export default OperationShell
