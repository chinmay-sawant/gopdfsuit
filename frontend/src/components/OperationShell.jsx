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
  <div className="glass-card operation-shell">
    <h3 className="operation-shell-title">
      <span className="operation-shell-icon" aria-hidden="true">{icon}</span>{title}
    </h3>
    {resultUrl ? (
      <div className="operation-shell-result">
        {stats}
        {kind === 'image' ? (
          <div className="operation-shell-image">
            <img src={resultUrl} alt={imageAlt} style={{ maxHeight: `${height}px` }} />
          </div>
        ) : (
          <iframe className="operation-shell-frame" src={resultUrl} style={{ height: `${height}px` }} title={title} />
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
      <div className="operation-shell-empty" style={{ minHeight: `${height}px` }}>
        <div>
          <div className="operation-shell-empty-icon" aria-hidden="true">{icon}</div>
          <p>{emptyTitle}</p>
          <span>{emptySubtitle}</span>
        </div>
      </div>
    )}
  </div>
)

export default OperationShell
