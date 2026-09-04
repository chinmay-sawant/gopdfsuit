import { useCallback, useState } from 'react'

/**
 * useToast owns toast list state so pages render a single ToastContainer.
 * Previously Editor kept its own showToast/removeToast pair and rendered
 * the toast list twice; use this seam instead.
 */
export const useToast = () => {
  const [toasts, setToasts] = useState([])

  const removeToast = useCallback((id) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id))
  }, [])

  const showToast = useCallback((message, type = 'success', duration = 3000) => {
    const id = Date.now() + Math.random()
    setToasts((prev) => [...prev, { id, message, type, duration }])
    return id
  }, [])

  return { toasts, showToast, removeToast }
}

export default useToast
