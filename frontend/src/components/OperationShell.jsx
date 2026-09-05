import { Download } from 'lucide-react'
import PdfPreview from './PdfPreview'

/**
 * OperationShell owns the shared result panel for PDF operations:
 * preview (shared PdfPreview iframe or image), optional stats row,
 * download action, and the empty placeholder shown before the first run.
 * Sizing lives in one place (.pdf-preview in index.css): the preview
 * fills the available width and viewport height on every tool.
 */
const OperationShell = ({
  resultUrl,
  title,
  icon,
  emptyTitle,
  emptySubtitle,
  onDownload,
  downloadLabel = 'Download',
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
        <PdfPreview url={resultUrl} title={title} kind={kind} imageAlt={imageAlt} />
        <button
          onClick={() => { if (onDownload) onDownload() }}
          disabled={isLoading}
          className="btn-glow operation-shell-download"
          style={{ width: 'auto', alignSelf: 'flex-start', marginTop: '1rem', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '0.6rem 1.1rem' }}
        >
          <Download size={16} />{isLoading && loadingLabel ? loadingLabel : downloadLabel}
        </button>
      </div>
    ) : (
      <div className="operation-shell-empty">
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
