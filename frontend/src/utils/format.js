export const formatFileSize = (bytes) => {
  if (!bytes) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export const resetDropStyles = (el) => {
  el.style.borderColor = 'rgba(255,255,255,0.15)'
  el.style.background = 'rgba(255,255,255,0.02)'
}

export default formatFileSize
