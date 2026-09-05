
import { lazy, Suspense } from 'react'
import { HashRouter as Router, Routes, Route, useLocation } from 'react-router-dom'
import AuthGuard from './components/AuthGuard'
import SiteFooter from './components/site/SiteFooter'
import SiteHeader from './components/site/SiteHeader'
import { isAuthRequired } from './utils/apiConfig'
import Home from './pages/Home'
import Editor from './pages/Editor'
import NotFound from './pages/NotFound'

const Viewer = lazy(() => import('./pages/Viewer'))
const Merge = lazy(() => import('./pages/Merge'))
const Split = lazy(() => import('./pages/Split'))
const Compress = lazy(() => import('./pages/Compress'))
const Filler = lazy(() => import('./pages/Filler'))
const HtmlToPdf = lazy(() => import('./pages/HtmlToPdf'))
const HtmlToImage = lazy(() => import('./pages/HtmlToImage'))
const Screenshots = lazy(() => import('./pages/Screenshots'))
const Comparison = lazy(() => import('./pages/Comparison'))
const Redaction = lazy(() => import('./pages/Redaction'))

const workspaceRoutes = new Set([
  '/viewer',
  '/merge',
  '/split',
  '/compress',
  '/filler',
  '/htmltopdf',
  '/htmltoimage',
  '/redact',
])

function AppLayout() {
  const { pathname } = useLocation()
  const isWorkspace = workspaceRoutes.has(pathname)

  // Public by default: Editor renders unwrapped. Only wrap in AuthGuard
  // when auth is explicitly opted in (VITE_IS_CLOUD_RUN=true).
  const EditorRoute = isAuthRequired() ? (
    <AuthGuard>
      <Editor />
    </AuthGuard>
  ) : (
    <Editor />
  )

  return (
    <div className={`app-shell${isWorkspace ? ' workspace-shell' : ''}`}>
      <SiteHeader />
      <main className={isWorkspace ? 'workspace-main' : ''} id="main-content" tabIndex="-1">
        <Suspense fallback={<div className="route-loading" role="status">Loading tool</div>}>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/viewer" element={<Viewer />} />
            <Route path="/editor" element={EditorRoute} />
            <Route path="/merge" element={<Merge />} />
            <Route path="/split" element={<Split />} />
            <Route path="/compress" element={<Compress />} />
            <Route path="/filler" element={<Filler />} />
            <Route path="/htmltopdf" element={<HtmlToPdf />} />
            <Route path="/htmltoimage" element={<HtmlToImage />} />
            <Route path="/screenshots" element={<Screenshots />} />
            <Route path="/comparison" element={<Comparison />} />
            <Route path="/redact" element={<Redaction />} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </Suspense>
      </main>
      {!isWorkspace && <SiteFooter />}
    </div>
  )
}

function App() {
  return (
    <Router>
      <AppLayout />
    </Router>
  )
}

export default App
