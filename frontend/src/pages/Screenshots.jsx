import React, { useState, useEffect, useRef } from 'react'
import { Camera, Loader, ChevronLeft, ChevronRight, Play, Pause } from 'lucide-react'

const Screenshots = () => {
  const [currentSlide, setCurrentSlide] = useState(0)
  const [loadingImages, setLoadingImages] = useState(new Set([1, 2, 3, 4, 5, 6, 7, 8]))
  const [autoPlay, setAutoPlay] = useState(true)
  const [slideDirection, setSlideDirection] = useState('right')
  const [isTransitioning, setIsTransitioning] = useState(false)
  const progressRef = useRef(null)
  const autoPlayRef = useRef(null)

  const screenshots = Array.from({ length: 8 }, (_, i) => i + 1)

  const handleImageLoad = (index) => {
    setLoadingImages(prev => {
      const newSet = new Set(prev)
      newSet.delete(index)
      return newSet
    })
  }

  const nextSlide = () => {
    if (isTransitioning) return
    setIsTransitioning(true)
    setSlideDirection('right')
    setTimeout(() => {
      setCurrentSlide((prev) => (prev + 1) % screenshots.length)
      setIsTransitioning(false)
    }, 250)
  }

  const prevSlide = () => {
    if (isTransitioning) return
    setIsTransitioning(true)
    setSlideDirection('left')
    setTimeout(() => {
      setCurrentSlide((prev) => (prev - 1 + screenshots.length) % screenshots.length)
      setIsTransitioning(false)
    }, 250)
  }

  const goToSlide = (index) => {
    if (isTransitioning || index === currentSlide) return
    setIsTransitioning(true)
    setSlideDirection(index > currentSlide ? 'right' : 'left')
    setTimeout(() => {
      setCurrentSlide(index)
      setIsTransitioning(false)
    }, 250)
  }

  // Auto-play functionality with progress bar
  useEffect(() => {
    if (!autoPlay) {
      if (progressRef.current) {
        progressRef.current.style.animationPlayState = 'paused'
      }
      return
    }

    if (progressRef.current) {
      progressRef.current.style.animationPlayState = 'running'
    }

    autoPlayRef.current = setInterval(() => {
      nextSlide()
    }, 4000)

    return () => {
      if (autoPlayRef.current) {
        clearInterval(autoPlayRef.current)
      }
    }
  }, [autoPlay, currentSlide, isTransitioning])

  // Reset progress bar animation when slide changes
  useEffect(() => {
    if (progressRef.current && autoPlay) {
      progressRef.current.style.animation = 'none'
      setTimeout(() => {
        progressRef.current.style.animation = 'progress 4s linear'
      }, 10)
    }
  }, [currentSlide, autoPlay])

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
  }, [isTransitioning])

  return (
    <div style={{
      minHeight: '100vh',
      padding: '2rem 0',
      background: 'linear-gradient(135deg, hsl(var(--background)) 0%, hsl(var(--muted)) 100%)',
      position: 'relative',
      overflow: 'hidden'
    }}>
      {/* Animated background elements */}
      <div style={{
        position: 'absolute',
        top: '10%',
        left: '10%',
        width: '200px',
        height: '200px',
        background: 'radial-gradient(circle, rgba(78, 205, 196, 0.1) 0%, transparent 70%)',
        borderRadius: '50%',
        animation: 'pulse 4s ease-in-out infinite'
      }} />
      <div style={{
        position: 'absolute',
        top: '60%',
        right: '15%',
        width: '150px',
        height: '150px',
        background: 'radial-gradient(circle, rgba(0, 122, 204, 0.1) 0%, transparent 70%)',
        borderRadius: '50%',
        animation: 'pulse 6s ease-in-out infinite reverse'
      }} />

      <div className="container" style={{ position: 'relative', zIndex: 1 }}>
        <div style={{
          textAlign: 'center',
          marginBottom: '3rem',
          animation: 'fadeIn 1s ease-out'
        }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '1rem'
          }}>
            <Camera size={32} style={{
              color: '#4ecdc4',
              animation: 'bounce 2s infinite'
            }} />
            <h1 style={{
              fontSize: '2.5rem',
              fontWeight: '800',
              color: 'hsl(var(--foreground))',
              margin: 0,
              background: 'linear-gradient(45deg, #4ecdc4, #007acc)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              backgroundClip: 'text'
            }}>
              Screenshots
            </h1>
          </div>
          <p style={{
            color: 'hsl(var(--muted-foreground))',
            fontSize: '1.1rem',
            maxWidth: '600px',
            margin: '0 auto 1rem',
            animation: 'fadeIn 1s ease-out 0.2s both'
          }}>
            Explore the GoPdfSuit interface through our collection of screenshots
          </p>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '1rem',
            animation: 'fadeIn 1s ease-out 0.4s both'
          }}>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '0.9rem',
              margin: 0
            }}>
              Use arrow keys to navigate • Press spacebar to {autoPlay ? 'pause' : 'play'} auto-play
            </p>
            <button
              onClick={() => setAutoPlay(prev => !prev)}
              style={{
                background: 'rgba(78, 205, 196, 0.1)',
                border: '1px solid rgba(78, 205, 196, 0.3)',
                borderRadius: '20px',
                padding: '0.5rem',
                cursor: 'pointer',
                color: '#4ecdc4',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                transition: 'all 0.3s ease'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(78, 205, 196, 0.2)'
                e.currentTarget.style.transform = 'scale(1.05)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'rgba(78, 205, 196, 0.1)'
                e.currentTarget.style.transform = 'scale(1)'
              }}
            >
              {autoPlay ? <Pause size={16} /> : <Play size={16} />}
            </button>
          </div>
        </div>

        {/* Slideshow Container */}
        <div style={{
          position: 'relative',
          maxWidth: '1200px',
          margin: '0 auto',
          borderRadius: '16px',
          overflow: 'hidden',
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(255, 255, 255, 0.1)',
          background: 'hsl(var(--card))',
          animation: 'fadeIn 1s ease-out 0.6s both'
        }}>
          {/* Progress Bar */}
          {autoPlay && (
            <div style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              height: '3px',
              background: 'rgba(255, 255, 255, 0.2)',
              zIndex: 30
            }}>
              <div
                ref={progressRef}
                style={{
                  height: '100%',
                  background: 'linear-gradient(90deg, #4ecdc4, #007acc)',
                  animation: 'progress 4s linear',
                  borderRadius: '0 3px 3px 0'
                }}
              />
            </div>
          )}

          {/* Main Image Display */}
          <div style={{
            position: 'relative',
            width: '100%',
            height: '70vh',
            minHeight: '500px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'linear-gradient(135deg, hsl(var(--muted)) 0%, rgba(78, 205, 196, 0.05) 100%)',
            overflow: 'hidden'
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
                zIndex: 10,
                animation: 'fadeIn 0.5s ease-out'
              }}>
                <div style={{
                  width: '80px',
                  height: '80px',
                  border: '4px solid rgba(78, 205, 196, 0.2)',
                  borderTop: '4px solid #4ecdc4',
                  borderRadius: '50%',
                  animation: 'spin 1s linear infinite'
                }} />
                <span style={{
                  fontSize: '1.1rem',
                  fontWeight: '500'
                }}>
                  Loading screenshot {screenshots[currentSlide]}...
                </span>
                <div style={{
                  width: '200px',
                  height: '4px',
                  background: 'rgba(255, 255, 255, 0.2)',
                  borderRadius: '2px',
                  overflow: 'hidden'
                }}>
                  <div style={{
                    height: '100%',
                    background: 'linear-gradient(90deg, #4ecdc4, #007acc)',
                    animation: 'progress 2s ease-in-out infinite',
                    borderRadius: '2px'
                  }} />
                </div>
              </div>
            )}
            <img
              key={currentSlide}
              src={`https://raw.githubusercontent.com/chinmay-sawant/gopdfsuit/master/screenshots/${screenshots[currentSlide]}.png`}
              alt={`Screenshot ${screenshots[currentSlide]}`}
              onLoad={() => handleImageLoad(screenshots[currentSlide])}
              onError={() => handleImageLoad(screenshots[currentSlide])}
              style={{
                maxWidth: '100%',
                maxHeight: '100%',
                objectFit: 'contain',
                display: loadingImages.has(screenshots[currentSlide]) ? 'none' : 'block',
                animation: loadingImages.has(screenshots[currentSlide]) ? 'none' : 'fadeIn 0.8s ease-out',
                filter: 'drop-shadow(0 10px 30px rgba(0, 0, 0, 0.2))'
              }}
            />
          </div>

          {/* Navigation Buttons */}
          <button
            onClick={prevSlide}
            disabled={isTransitioning}
            style={{
              position: 'absolute',
              left: '1.5rem',
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'rgba(0, 0, 0, 0.6)',
              backdropFilter: 'blur(10px)',
              color: 'white',
              border: 'none',
              borderRadius: '50%',
              width: '60px',
              height: '60px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: isTransitioning ? 'not-allowed' : 'pointer',
              transition: 'all 0.3s ease',
              zIndex: 20,
              opacity: isTransitioning ? 0.5 : 1,
              boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)'
            }}
            onMouseEnter={(e) => {
              if (!isTransitioning) {
                e.currentTarget.style.background = 'rgba(0, 0, 0, 0.8)'
                e.currentTarget.style.transform = 'translateY(-50%) scale(1.1)'
                e.currentTarget.style.boxShadow = '0 12px 40px rgba(0, 0, 0, 0.4)'
              }
            }}
            onMouseLeave={(e) => {
              if (!isTransitioning) {
                e.currentTarget.style.background = 'rgba(0, 0, 0, 0.6)'
                e.currentTarget.style.transform = 'translateY(-50%) scale(1)'
                e.currentTarget.style.boxShadow = '0 8px 32px rgba(0, 0, 0, 0.3)'
              }
            }}
          >
            <ChevronLeft size={28} />
          </button>

          <button
            onClick={nextSlide}
            disabled={isTransitioning}
            style={{
              position: 'absolute',
              right: '1.5rem',
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'rgba(0, 0, 0, 0.6)',
              backdropFilter: 'blur(10px)',
              color: 'white',
              border: 'none',
              borderRadius: '50%',
              width: '60px',
              height: '60px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: isTransitioning ? 'not-allowed' : 'pointer',
              transition: 'all 0.3s ease',
              zIndex: 20,
              opacity: isTransitioning ? 0.5 : 1,
              boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)'
            }}
            onMouseEnter={(e) => {
              if (!isTransitioning) {
                e.currentTarget.style.background = 'rgba(0, 0, 0, 0.8)'
                e.currentTarget.style.transform = 'translateY(-50%) scale(1.1)'
                e.currentTarget.style.boxShadow = '0 12px 40px rgba(0, 0, 0, 0.4)'
              }
            }}
            onMouseLeave={(e) => {
              if (!isTransitioning) {
                e.currentTarget.style.background = 'rgba(0, 0, 0, 0.6)'
                e.currentTarget.style.transform = 'translateY(-50%) scale(1)'
                e.currentTarget.style.boxShadow = '0 8px 32px rgba(0, 0, 0, 0.3)'
              }
            }}
          >
            <ChevronRight size={28} />
          </button>

          {/* Slide Indicators */}
          <div style={{
            position: 'absolute',
            bottom: '2rem',
            left: '50%',
            transform: 'translateX(-50%)',
            display: 'flex',
            gap: '0.75rem',
            zIndex: 20,
            animation: 'fadeIn 1s ease-out 0.8s both'
          }}>
            {screenshots.map((_, index) => (
              <button
                key={index}
                onClick={() => goToSlide(index)}
                disabled={isTransitioning}
                style={{
                  width: '14px',
                  height: '14px',
                  borderRadius: '50%',
                  border: 'none',
                  background: index === currentSlide
                    ? 'linear-gradient(45deg, #4ecdc4, #007acc)'
                    : 'rgba(255, 255, 255, 0.4)',
                  cursor: isTransitioning ? 'not-allowed' : 'pointer',
                  transition: 'all 0.3s ease',
                  boxShadow: index === currentSlide
                    ? '0 0 20px rgba(78, 205, 196, 0.5)'
                    : 'none',
                  transform: index === currentSlide ? 'scale(1.2)' : 'scale(1)'
                }}
                onMouseEnter={(e) => {
                  if (!isTransitioning && index !== currentSlide) {
                    e.currentTarget.style.transform = 'scale(1.1)'
                    e.currentTarget.style.background = 'rgba(255, 255, 255, 0.7)'
                  }
                }}
                onMouseLeave={(e) => {
                  if (!isTransitioning && index !== currentSlide) {
                    e.currentTarget.style.transform = 'scale(1)'
                    e.currentTarget.style.background = 'rgba(255, 255, 255, 0.4)'
                  }
                }}
              />
            ))}
          </div>

          {/* Slide Counter */}
          <div style={{
            position: 'absolute',
            top: '1.5rem',
            right: '1.5rem',
            background: 'rgba(0, 0, 0, 0.8)',
            backdropFilter: 'blur(10px)',
            color: 'white',
            padding: '0.75rem 1.25rem',
            borderRadius: '25px',
            fontSize: '0.95rem',
            fontWeight: '600',
            zIndex: 20,
            animation: 'fadeIn 1s ease-out 0.8s both',
            boxShadow: '0 4px 20px rgba(0, 0, 0, 0.3)'
          }}>
            {currentSlide + 1} / {screenshots.length}
          </div>
        </div>

        {/* Enhanced Thumbnail Strip */}
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          gap: '1rem',
          marginTop: '3rem',
          flexWrap: 'wrap',
          animation: 'fadeIn 1s ease-out 1s both'
        }}>
          {screenshots.map((num, index) => (
            <button
              key={num}
              onClick={() => goToSlide(index)}
              disabled={isTransitioning}
              style={{
                width: '90px',
                height: '68px',
                border: index === currentSlide
                  ? '3px solid #4ecdc4'
                  : '3px solid transparent',
                borderRadius: '10px',
                overflow: 'hidden',
                cursor: isTransitioning ? 'not-allowed' : 'pointer',
                opacity: index === currentSlide ? 1 : 0.7,
                transition: 'all 0.4s ease',
                transform: index === currentSlide ? 'scale(1.05)' : 'scale(1)',
                boxShadow: index === currentSlide
                  ? '0 8px 25px rgba(78, 205, 196, 0.3)'
                  : '0 4px 15px rgba(0, 0, 0, 0.1)',
                position: 'relative'
              }}
              onMouseEnter={(e) => {
                if (!isTransitioning) {
                  e.currentTarget.style.transform = index === currentSlide ? 'scale(1.1)' : 'scale(1.08)'
                  e.currentTarget.style.opacity = '1'
                  e.currentTarget.style.boxShadow = '0 8px 25px rgba(78, 205, 196, 0.2)'
                }
              }}
              onMouseLeave={(e) => {
                if (!isTransitioning) {
                  e.currentTarget.style.transform = index === currentSlide ? 'scale(1.05)' : 'scale(1)'
                  e.currentTarget.style.opacity = index === currentSlide ? 1 : 0.7
                  e.currentTarget.style.boxShadow = index === currentSlide
                    ? '0 8px 25px rgba(78, 205, 196, 0.3)'
                    : '0 4px 15px rgba(0, 0, 0, 0.1)'
                }
              }}
            >
              <img
                src={`https://raw.githubusercontent.com/chinmay-sawant/gopdfsuit/master/screenshots/${num}.png`}
                alt={`Thumbnail ${num}`}
                style={{
                  width: '100%',
                  height: '100%',
                  objectFit: 'cover',
                  transition: 'all 0.3s ease'
                }}
              />
              {index === currentSlide && (
                <div style={{
                  position: 'absolute',
                  top: '4px',
                  right: '4px',
                  width: '8px',
                  height: '8px',
                  background: '#4ecdc4',
                  borderRadius: '50%',
                  animation: 'pulse 2s infinite'
                }} />
              )}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Screenshots