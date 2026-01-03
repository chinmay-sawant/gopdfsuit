import React from 'react'
import { HashRouter as Router, Routes, Route } from 'react-router-dom'
import Navbar from './components/Navbar'
import AuthGuard from './components/AuthGuard'
import { isAuthRequired } from './utils/apiConfig'
import Home from './pages/Home'
import Editor from './pages/Editor'
import Viewer from './pages/Viewer'
import Merge from './pages/Merge'
import Filler from './pages/Filler'
import HtmlToPdf from './pages/HtmlToPdf'
import HtmlToImage from './pages/HtmlToImage'
import Screenshots from './pages/Screenshots'
import Comparison from './pages/Comparison'

function App() {
  // If auth is required (Cloud Run deployment), wrap entire app with AuthGuard
  // Otherwise, show pages directly
  const AppContent = () => (
    <>
      <Navbar />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/viewer" element={<Viewer />} />
        <Route path="/editor" element={<Editor />} />
        <Route path="/merge" element={<Merge />} />
        <Route path="/filler" element={<Filler />} />
        <Route path="/htmltopdf" element={<HtmlToPdf />} />
        <Route path="/htmltoimage" element={<HtmlToImage />} />
        <Route path="/screenshots" element={<Screenshots />} />
        <Route path="/comparison" element={<Comparison />} />
      </Routes>
    </>
  )

  return (
    <Router>
      <div className="App">
        {isAuthRequired() ? (
          <AuthGuard>
            <AppContent />
          </AuthGuard>
        ) : (
          <AppContent />
        )}
      </div>
    </Router>
  )
}

export default App