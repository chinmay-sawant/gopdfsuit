import { useId, useRef, useState } from 'react'
import { Upload } from 'lucide-react'

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
}) => {
  const fileInputRef = useRef(null)
  const [isDragging, setIsDragging] = useState(false)
  const helpId = useId()

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
      <button
        aria-describedby={helpId}
        className={`file-dropzone${isDragging ? ' is-dragging' : ''}`}
        onClick={() => { if (!disabled) fileInputRef.current?.click() }}
        disabled={disabled}
        onDragOver={(event) => { event.preventDefault(); if (!disabled) setIsDragging(true) }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={(event) => { event.preventDefault(); setIsDragging(false); if (!disabled) emit(event.dataTransfer.files) }}
        type="button"
      >
        <Upload aria-hidden="true" size={28} />
        <strong>{title}</strong>
        <span id={helpId}>{subtitle}</span>
      </button>
    </>
  )
}

export default FileDropzone
