import React from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Navbar from './components/Navbar'
import Home from './pages/Home'
import Editor from './pages/Editor'
import Viewer from './pages/Viewer'
import Merge from './pages/Merge'
import Filler from './pages/Filler'
import HtmlToPdf from './pages/HtmlToPdf'
import HtmlToImage from './pages/HtmlToImage'
import Screenshots from './pages/Screenshots'

function App() {
  return (
    <Router>
      <div className="App">
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
        </Routes>
      </div>
    </Router>
  )
}

export default App