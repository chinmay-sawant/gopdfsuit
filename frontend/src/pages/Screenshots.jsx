import React, { useState, useEffect } from 'react'
import { Camera, Loader, ChevronLeft, ChevronRight } from 'lucide-react'

const Screenshots = () => {
  const [currentSlide, setCurrentSlide] = useState(0)
  const [loadingImages, setLoadingImages] = useState(new Set([1, 2, 3, 4, 5, 6, 7, 8]))
  const [autoPlay, setAutoPlay] = useState(true)

  const screenshots = Array.from({ length: 8 }, (_, i) => i + 1)

  const handleImageLoad = (index) => {
    setLoadingImages(prev => {
      const newSet = new Set(prev)
      newSet.delete(index)
      return newSet
    })
  }

  const nextSlide = () => {
    setCurrentSlide((prev) => (prev + 1) % screenshots.length)
  }

  const prevSlide = () => {
    setCurrentSlide((prev) => (prev - 1 + screenshots.length) % screenshots.length)
  }

  const goToSlide = (index) => {
    setCurrentSlide(index)
  }

  // Auto-play functionality
  useEffect(() => {
    if (!autoPlay) return

    const interval = setInterval(() => {
      nextSlide()
    }, 4000) // Change slide every 4 seconds

    return () => clearInterval(interval)
  }, [autoPlay, currentSlide])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyPress = (e) => {
      if (e.key === 'ArrowLeft') {
        prevSlide()
      } else if (e.key === 'ArrowRight') {
        nextSlide()
      } else if (e.key === ' ') {
        e.preventDefault()
        setAutoPlay(prev => !prev)
      }
    }

    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [])

  return (
    <div style={{ minHeight: '100vh', padding: '2rem 0', background: 'hsl(var(--background))' }}>
      <div className="container">
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <div style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '1rem'
          }}>
            <Camera size={32} style={{ color: '#4ecdc4' }} />
            <h1 style={{ 
              fontSize: '2.5rem',
              fontWeight: '800',
              color: 'hsl(var(--foreground))',
              margin: 0
            }}>
              Screenshots
            </h1>
          </div>
          <p style={{ 
            color: 'hsl(var(--muted-foreground))',
            fontSize: '1.1rem',
            maxWidth: '600px',
            margin: '0 auto 1rem'
          }}>
            Explore the GoPdfSuit interface through our collection of screenshots
          </p>
          <p style={{ 
            color: 'hsl(var(--muted-foreground))',
            fontSize: '0.9rem',
            margin: 0
          }}>
            Use arrow keys to navigate • Press spacebar to {autoPlay ? 'pause' : 'play'} auto-play
          </p>
        </div>

        {/* Slideshow Container */}
        <div style={{
          position: 'relative',
          maxWidth: '1200px',
          margin: '0 auto',
          borderRadius: '12px',
          overflow: 'hidden',
          boxShadow: '0 10px 40px rgba(0, 0, 0, 0.1)',
          background: 'hsl(var(--card))'
        }}>
          {/* Main Image Display */}
          <div style={{
            position: 'relative',
            width: '100%',
            height: '70vh',
            minHeight: '500px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'hsl(var(--muted))'
          }}>
            {loadingImages.has(screenshots[currentSlide]) && (
              <div style={{
                position: 'absolute',
                top: '50%',
                left: '50%',
                transform: 'translate(-50%, -50%)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: '1rem',
                color: 'hsl(var(--muted-foreground))',
                zIndex: 10
              }}>
                <Loader size={48} className="animate-spin" />
                <span style={{ fontSize: '1.1rem' }}>Loading screenshot {screenshots[currentSlide]}...</span>
              </div>
            )}
            <img
              src={`https://raw.githubusercontent.com/chinmay-sawant/gopdfsuit/fix-35-newwebsite/screenshots/${screenshots[currentSlide]}.png`}
              alt={`Screenshot ${screenshots[currentSlide]}`}
              onLoad={() => handleImageLoad(screenshots[currentSlide])}
              onError={() => handleImageLoad(screenshots[currentSlide])}
              style={{
                maxWidth: '100%',
                maxHeight: '100%',
                objectFit: 'contain',
                display: loadingImages.has(screenshots[currentSlide]) ? 'none' : 'block'
              }}
            />
          </div>

          {/* Navigation Buttons */}
          <button
            onClick={prevSlide}
            style={{
              position: 'absolute',
              left: '1rem',
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'rgba(0, 0, 0, 0.5)',
              color: 'white',
              border: 'none',
              borderRadius: '50%',
              width: '50px',
              height: '50px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              transition: 'all 0.3s ease',
              zIndex: 20
            }}
            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(0, 0, 0, 0.8)'}
            onMouseLeave={(e) => e.currentTarget.style.background = 'rgba(0, 0, 0, 0.5)'}
          >
            <ChevronLeft size={24} />
          </button>

          <button
            onClick={nextSlide}
            style={{
              position: 'absolute',
              right: '1rem',
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'rgba(0, 0, 0, 0.5)',
              color: 'white',
              border: 'none',
              borderRadius: '50%',
              width: '50px',
              height: '50px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              transition: 'all 0.3s ease',
              zIndex: 20
            }}
            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(0, 0, 0, 0.8)'}
            onMouseLeave={(e) => e.currentTarget.style.background = 'rgba(0, 0, 0, 0.5)'}
          >
            <ChevronRight size={24} />
          </button>

          {/* Slide Indicators */}
          <div style={{
            position: 'absolute',
            bottom: '1rem',
            left: '50%',
            transform: 'translateX(-50%)',
            display: 'flex',
            gap: '0.5rem',
            zIndex: 20
          }}>
            {screenshots.map((_, index) => (
              <button
                key={index}
                onClick={() => goToSlide(index)}
                style={{
                  width: '12px',
                  height: '12px',
                  borderRadius: '50%',
                  border: 'none',
                  background: index === currentSlide ? '#4ecdc4' : 'rgba(255, 255, 255, 0.5)',
                  cursor: 'pointer',
                  transition: 'all 0.3s ease'
                }}
              />
            ))}
          </div>

          {/* Slide Counter */}
          <div style={{
            position: 'absolute',
            top: '1rem',
            right: '1rem',
            background: 'rgba(0, 0, 0, 0.7)',
            color: 'white',
            padding: '0.5rem 1rem',
            borderRadius: '20px',
            fontSize: '0.9rem',
            zIndex: 20
          }}>
            {currentSlide + 1} / {screenshots.length}
          </div>
        </div>

        {/* Thumbnail Strip (Optional) */}
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          gap: '0.5rem',
          marginTop: '2rem',
          flexWrap: 'wrap'
        }}>
          {screenshots.map((num, index) => (
            <button
              key={num}
              onClick={() => goToSlide(index)}
              style={{
                width: '80px',
                height: '60px',
                border: index === currentSlide ? '2px solid #4ecdc4' : '2px solid transparent',
                borderRadius: '6px',
                overflow: 'hidden',
                cursor: 'pointer',
                opacity: index === currentSlide ? 1 : 0.6,
                transition: 'all 0.3s ease'
              }}
            >
              <img
                src={`https://raw.githubusercontent.com/chinmay-sawant/gopdfsuit/fix-35-newwebsite/screenshots/${num}.png`}
                alt={`Thumbnail ${num}`}
                style={{
                  width: '100%',
                  height: '100%',
                  objectFit: 'cover'
                }}
              />
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Screenshots