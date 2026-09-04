import { useRef } from 'react'
import { Upload } from 'lucide-react'
import { resetDropStyles } from '../utils/format'

/**
 * FileDropzone owns the shared click-or-drag file input every op page
 * pasted by hand. Calls onFiles(File[]) with the dropped/selected files;
 * type filtering stays with the caller (single vs multi, pdf vs json).
 */
const FileDropzone = ({
  onFiles,
  accept = '.pdf,application/pdf',
  multiple = false,
  title = 'Click to upload or drag & drop',
  subtitle = 'Select a PDF file',
  disabled = false,
  compact = false,
}) => {
  const fileInputRef = useRef(null)

  const emit = (files) => {
    if (disabled) return
    const list = Array.from(files || [])
    if (list.length > 0) onFiles(list)
  }

  return (
    <>
      <input
        ref={fileInputRef}
        type="file"
        accept={accept}
        multiple={multiple}
        onChange={(event) => { emit(event.target.files); event.target.value = '' }}
        disabled={disabled}
        style={{ display: 'none' }}
      />
      <div
        onClick={() => { if (!disabled) fileInputRef.current?.click() }}
        style={{ border: '2px dashed rgba(255,255,255,0.15)', borderRadius: '8px', padding: compact ? '2rem' : '3rem 2rem', textAlign: 'center', cursor: disabled ? 'not-allowed' : 'pointer', transition: 'all 0.3s ease', marginBottom: '2rem', background: 'rgba(255,255,255,0.02)', opacity: disabled ? 0.6 : 1 }}
        onDragOver={(e) => { e.preventDefault(); if (disabled) return; e.currentTarget.style.borderColor = '#4ecdc4'; e.currentTarget.style.background = 'rgba(78,205,196,0.1)' }}
        onDragLeave={(e) => { resetDropStyles(e.currentTarget) }}
        onDrop={(e) => { e.preventDefault(); resetDropStyles(e.currentTarget); if (disabled) return; emit(e.dataTransfer.files) }}
      >
        <div className="feature-icon-box teal" style={{ width: '56px', height: '56px', margin: '0 auto 1rem', opacity: 0.6 }}><Upload size={28} /></div>
        <p style={{ color: 'hsl(var(--foreground))', marginBottom: '0.5rem', fontSize: '1.1rem', fontWeight: '600' }}>{title}</p>
        <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.9rem', marginBottom: 0 }}>{subtitle}</p>
      </div>
    </>
  )
}

export default FileDropzone
