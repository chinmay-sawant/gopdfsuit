import { useState, useEffect } from 'react'

export function useScrollAnimation() {
  const [isVisible, setIsVisible] = useState({})

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsVisible(prev => ({ ...prev, [entry.target.id]: true }))
          }
        })
      },
      { threshold: 0.1 }
    )

    const sections = document.querySelectorAll('[id^="section-"]')
    sections.forEach((section) => observer.observe(section))

    return () => observer.disconnect()
  }, [])

  return isVisible
}
