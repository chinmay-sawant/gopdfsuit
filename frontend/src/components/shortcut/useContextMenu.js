import { useState, useEffect, useCallback } from 'react'

export default function useContextMenu() {
  const [menuState, setMenuState] = useState({
    visible: false,
    x: 0,
    y: 0,
    targetType: null, // 'canvas' | 'title' | 'table' | 'cell' | 'spacer' | 'image' | 'footer'
    targetData: {}    // { elementId, rowIdx?, colIdx?, element?, index? }
  })

  const showMenu = useCallback((e, targetType, targetData = {}) => {
    e.preventDefault()
    e.stopPropagation()

    setMenuState({ visible: true, x: e.clientX, y: e.clientY, targetType, targetData })
  }, [])

  const hideMenu = useCallback(() => {
    setMenuState(prev => ({ ...prev, visible: false }))
  }, [])

  // Auto-dismiss on outside click, scroll, or Escape
  useEffect(() => {
    if (!menuState.visible) return

    const handleDismiss = () => hideMenu()
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') hideMenu()
    }

    // Use setTimeout to avoid the same click that opened the menu from closing it
    const timer = setTimeout(() => {
      document.addEventListener('click', handleDismiss)
      document.addEventListener('scroll', handleDismiss, true)
      window.addEventListener('keydown', handleKeyDown)
    }, 0)

    return () => {
      clearTimeout(timer)
      document.removeEventListener('click', handleDismiss)
      document.removeEventListener('scroll', handleDismiss, true)
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [menuState.visible, hideMenu])

  return { menuState, showMenu, hideMenu }
}
