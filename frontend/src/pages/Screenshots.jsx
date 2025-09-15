import React, { useState } from 'react'
import { Camera, Loader } from 'lucide-react'

const Screenshots = () => {
  const [loadingImages, setLoadingImages] = useState(new Set([1, 2, 3, 4, 5, 6, 7, 8]))

  const handleImageLoad = (index) => {
    setLoadingImages(prev => {
      const newSet = new Set(prev)
      newSet.delete(index)
      return newSet
    })
  }

  const screenshots = Array.from({ length: 8 }, (_, i) => i + 1)

  return (
    <div style={{ minHeight: '100vh', padding: '2rem 0' }}>
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
            margin: '0 auto'
          }}>
            Explore the GoPdfSuit interface through our collection of screenshots
          </p>
        </div>

        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))',
          gap: '2rem',
          alignItems: 'start'
        }}>
          {screenshots.map((num) => (
            <div key={num} className="card" style={{ 
              overflow: 'hidden',
              position: 'relative',
              minHeight: '300px'
            }}>
              <div style={{ 
                position: 'relative',
                width: '100%',
                height: 'auto',
                minHeight: '250px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'hsl(var(--muted))',
                borderRadius: '8px 8px 0 0'
              }}>
                {loadingImages.has(num) && (
                  <div style={{
                    position: 'absolute',
                    top: '50%',
                    left: '50%',
                    transform: 'translate(-50%, -50%)',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: '1rem',
                    color: 'hsl(var(--muted-foreground))'
                  }}>
                    <Loader size={32} className="animate-spin" />
                    <span style={{ fontSize: '0.9rem' }}>Loading screenshot {num}...</span>
                  </div>
                )}
                <img
                  src={`https://raw.githubusercontent.com/chinmay-sawant/gopdfsuit/fix-35-newwebsite/screenshots/${num}.png`}
                  alt={`Screenshot ${num}`}
                  onLoad={() => handleImageLoad(num)}
                  onError={() => handleImageLoad(num)}
                  style={{
                    width: '100%',
                    height: 'auto',
                    display: loadingImages.has(num) ? 'none' : 'block',
                    borderRadius: '8px 8px 0 0'
                  }}
                />
              </div>
              <div style={{ padding: '1rem' }}>
                <h3 style={{ 
                  margin: 0, 
                  color: 'hsl(var(--foreground))',
                  fontSize: '1.1rem'
                }}>
                  Screenshot {num}
                </h3>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Screenshots