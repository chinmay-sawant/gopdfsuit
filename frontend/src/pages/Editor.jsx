import React from 'react'
import { Edit, Wrench } from 'lucide-react'

const Editor = () => {
  return (
    <div style={{ padding: '2rem 0', minHeight: '100vh' }}>
      <div className="container">
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <h1 style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            gap: '1rem',
            marginBottom: '1rem',
            color: 'white',
          }}>
            <Edit size={40} />
            PDF Template Editor
          </h1>
          <p style={{ color: 'rgba(255, 255, 255, 0.8)', fontSize: '1.1rem' }}>
            Drag-and-drop visual template builder with live JSON generation
          </p>
        </div>

        <div className="card" style={{ textAlign: 'center', padding: '4rem 2rem' }}>
          <div style={{ fontSize: '4rem', marginBottom: '2rem' }}>🚧</div>
          <h2 style={{ color: 'white', marginBottom: '1rem' }}>Under Construction</h2>
          <p style={{ color: 'rgba(255, 255, 255, 0.8)', marginBottom: '2rem' }}>
            The visual PDF template editor is being migrated to React. This will feature:
          </p>
          
          <div className="grid grid-2" style={{ textAlign: 'left', marginTop: '2rem' }}>
            <div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Wrench size={16} />
                Features
              </h4>
              <ul style={{ color: 'rgba(255, 255, 255, 0.8)', lineHeight: 1.8 }}>
                <li>🎨 Drag-and-Drop Interface</li>
                <li>📋 Real-time JSON Generation</li>
                <li>🔧 Component Properties Panel</li>
                <li>📄 Live PDF Preview</li>
              </ul>
            </div>
            <div>
              <h4 style={{ color: '#4ecdc4', marginBottom: '1rem' }}>Coming Soon</h4>
              <ul style={{ color: 'rgba(255, 255, 255, 0.8)', lineHeight: 1.8 }}>
                <li>💾 Template Loading & Saving</li>
                <li>📱 Responsive Design</li>
                <li>🎨 Multiple Themes</li>
                <li>🧩 Component Library</li>
              </ul>
            </div>
          </div>

          <div style={{ marginTop: '3rem' }}>
            <p style={{ color: 'rgba(255, 255, 255, 0.6)', marginBottom: '1rem' }}>
              For now, you can use the PDF Viewer to work with JSON templates directly.
            </p>
            <a href="/viewer" className="btn btn-primary">
              Go to PDF Viewer
            </a>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Editor