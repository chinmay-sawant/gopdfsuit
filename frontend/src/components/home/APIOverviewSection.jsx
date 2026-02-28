import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/react'
import { Zap, Globe } from 'lucide-react'

const restEndpoints = [
  { method: 'POST', path: '/api/v1/generate/template-pdf', desc: 'Generate PDF' },
  { method: 'POST', path: '/api/v1/merge', desc: 'Merge PDFs' },
  { method: 'POST', path: '/api/v1/split', desc: 'Split PDFs' },
  { method: 'POST', path: '/api/v1/fill', desc: 'Fill forms' },
  { method: 'GET', path: '/api/v1/template-data', desc: 'Get template data' },
  { method: 'GET', path: '/api/v1/fonts', desc: 'List fonts' },
  { method: 'POST', path: '/api/v1/htmltopdf', desc: 'HTML to PDF' },
  { method: 'POST', path: '/api/v1/htmltoimage', desc: 'HTML to Image' },
  { method: 'POST', path: '/api/v1/redact/page-info', desc: 'Get page info' },
  { method: 'POST', path: '/api/v1/redact/search', desc: 'Search text for redaction' },
  { method: 'POST', path: '/api/v1/redact/apply', desc: 'Apply redactions' },
]

const webRoutes = [
  { path: '/viewer', desc: 'PDF Viewer & Template Processor' },
  { path: '/editor', desc: 'Drag-and-drop Template Editor' },
  { path: '/merge', desc: 'PDF Merge Interface' },
  { path: '/split', desc: 'PDF Split Interface' },
  { path: '/filler', desc: 'PDF Form Filler' },
  { path: '/htmltopdf', desc: 'HTML to PDF Converter' },
  { path: '/htmltoimage', desc: 'HTML to Image Converter' },
  { path: '/screenshots', desc: 'Screenshots Page' },
  { path: '/comparison', desc: 'Feature Comparison' },
  { path: '/redact', desc: 'PDF Redaction Tool' },
]

const APIOverviewSection = ({ isVisible }) => {
  const visible = isVisible['section-api']

  return (
    <section id="section-api" style={{ padding: '5rem 0' }}>
      <div className="container">
        <div
          className={`text-center animate-fadeInUp stagger-animation ${visible ? 'visible' : ''}`}
          style={{ marginBottom: '3rem' }}
        >
          <h2 className="gradient-text section-heading">
            API Endpoints
          </h2>
          <p className="section-subheading">
            RESTful API for seamless integration
          </p>
        </div>

        <div
          className={`glass-card animate-fadeInScale stagger-animation ${visible ? 'visible' : ''}`}
          style={{ padding: '2rem' }}
        >
          <TabGroup>
            <TabList className="api-tab-list">
              <Tab className="api-tab-button">
                <Zap size={18} />
                REST API
              </Tab>
              <Tab className="api-tab-button">
                <Globe size={18} />
                Web Interfaces
              </Tab>
            </TabList>

            <TabPanels>
              <TabPanel>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                  {restEndpoints.map((api, index) => (
                    <div key={index} className="api-endpoint-row">
                      <span className={`api-method-badge ${api.method.toLowerCase()}`}>
                        {api.method}
                      </span>
                      <code className="api-endpoint-path">{api.path}</code>
                      <span className="api-endpoint-desc">{api.desc}</span>
                    </div>
                  ))}
                </div>
              </TabPanel>

              <TabPanel>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                  {webRoutes.map((route, index) => (
                    <div key={index} className="api-endpoint-row">
                      <span className="api-method-badge web">
                        GET
                      </span>
                      <code className="api-endpoint-path">{route.path}</code>
                      <span className="api-endpoint-desc">{route.desc}</span>
                    </div>
                  ))}
                </div>
              </TabPanel>
            </TabPanels>
          </TabGroup>
        </div>
      </div>
    </section>
  )
}

export default APIOverviewSection
