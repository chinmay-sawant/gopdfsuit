import { useEffect } from 'react'

/**
 * useEditorShortcuts owns the Editor keyboard map (extracted from
 * Editor.jsx so the page keeps document state, not input wiring):
 * Ctrl/Cmd+S saves JSON, Delete removes the selection, Ctrl+C/X/V/D
 * clipboard ops, Ctrl+B/I/U style bits, Alt+ArrowUp/Down reordering.
 * Typing targets (input/textarea/select) only keep Ctrl+S.
 */
export function useEditorShortcuts({
  selectedId,
  selectedCell,
  clipboard,
  allElements,
  components,
  title,
  footer,
  jsonText,
  onSaveJson,
  onDelete,
  onCopy,
  onCut,
  onPaste,
  onDuplicate,
  onToggleStyle,
  onToggleCellStyle,
  onMove,
} = {}) {
  useEffect(() => {
    const handleKeyDown = (e) => {
      const tag = e.target.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') {
        if ((e.metaKey || e.ctrlKey) && e.key === 's') {
          e.preventDefault()
          onSaveJson()
        }
        return
      }

      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        onSaveJson()
        return
      }

      if (!selectedId) return

      const ctrl = e.metaKey || e.ctrlKey

      if (e.key === 'Delete' || e.key === 'Backspace') {
        e.preventDefault()
        onDelete(selectedId)
      } else if (ctrl && e.key === 'c') {
        e.preventDefault()
        onCopy(selectedId)
      } else if (ctrl && e.key === 'x') {
        e.preventDefault()
        onCut(selectedId)
      } else if (ctrl && e.key === 'v') {
        e.preventDefault()
        onPaste(selectedId)
      } else if (ctrl && e.key === 'd') {
        e.preventDefault()
        onDuplicate(selectedId)
      } else if (ctrl && e.key === 'b') {
        e.preventDefault()
        if (selectedCell) {
          onToggleCellStyle(selectedCell.elementId, selectedCell.rowIdx, selectedCell.colIdx, 0)
        } else {
          onToggleStyle(selectedId, 0)
        }
      } else if (ctrl && e.key === 'i') {
        e.preventDefault()
        if (selectedCell) {
          onToggleCellStyle(selectedCell.elementId, selectedCell.rowIdx, selectedCell.colIdx, 1)
        } else {
          onToggleStyle(selectedId, 1)
        }
      } else if (ctrl && e.key === 'u') {
        e.preventDefault()
        if (selectedCell) {
          onToggleCellStyle(selectedCell.elementId, selectedCell.rowIdx, selectedCell.colIdx, 2)
        } else {
          onToggleStyle(selectedId, 2)
        }
      } else if (e.altKey && e.key === 'ArrowUp') {
        e.preventDefault()
        const el = allElements.find(el => el.id === selectedId)
        if (el && el.type !== 'title' && el.type !== 'footer') {
          const idx = parseInt(selectedId.split('-')[1])
          onMove(idx, 'up')
        }
      } else if (e.altKey && e.key === 'ArrowDown') {
        e.preventDefault()
        const el = allElements.find(el => el.id === selectedId)
        if (el && el.type !== 'title' && el.type !== 'footer') {
          const idx = parseInt(selectedId.split('-')[1])
          onMove(idx, 'down')
        }
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [jsonText, selectedId, selectedCell, clipboard, allElements, components, title, footer])
}

export default useEditorShortcuts
