/**
 * PdfPreview is the single shared result preview for every tool.
 * It fills 100% of the available width and stretches to the viewport
 * height (see .pdf-preview in index.css), so Merge, Split, Compress,
 * Filler, Viewer, Editor, and HTML-convert all render the same frame
 * instead of each page owning its own pixel height.
 */
const PdfPreview = ({ url, title, kind = 'pdf', imageAlt = 'Generated output' }) => {
  if (!url) return null
  if (kind === 'image') {
    return (
      <div className="pdf-preview" data-kind="image">
        <img src={url} alt={imageAlt} />
      </div>
    )
  }
  return (
    <div className="pdf-preview" data-kind="pdf">
      <iframe src={url} title={title} loading="lazy" />
    </div>
  )
}

export default PdfPreview
