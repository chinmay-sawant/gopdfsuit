import { ArrowRight, FileCode2, ShieldCheck } from 'lucide-react'
import { Link } from 'react-router-dom'
import { toolGroups } from '../content/tools'

const Home = () => {
  return (
    <div className="home-page">
      <section className="home-hero">
        <div className="home-hero-copy">
          <h1>PDF work, without the detour.</h1>
          <p className="home-lede">Build templates and work with PDF files. Render HTML when that is where you start.</p>
          <div className="home-actions">
            <Link className="button button-primary" to="/editor">Open the editor <ArrowRight aria-hidden="true" size={17} /></Link>
            <Link className="button button-secondary" to="/viewer">Preview a template</Link>
          </div>
        </div>
        <aside className="home-note" aria-label="How the tools work">
          <FileCode2 aria-hidden="true" size={24} strokeWidth={1.6} />
          <h2>Made for a real document flow.</h2>
          <p>Each workspace keeps input, options, output, and download in one place.</p>
          <p>When a tool can process a file in the browser, it says so. When it needs the service, it asks before uploading.</p>
        </aside>
      </section>

      <section className="tool-catalogue" aria-labelledby="tool-catalogue-title">
        <div className="section-heading-row">
          <div>
            <h2 id="tool-catalogue-title">Start with the document in front of you.</h2>
          </div>
          <Link className="text-link" to="/comparison">See the product map <ArrowRight aria-hidden="true" size={16} /></Link>
        </div>
        {toolGroups.map((group) => (
          <section className="tool-group" key={group.title} aria-labelledby={`${group.title}-tools`}>
            <h3 id={`${group.title}-tools`}>{group.title}</h3>
            <div className="tool-card-grid">
              {group.items.map(({ to, label, description, icon: Icon }) => (
                <Link className="tool-card" key={`${to}-${label}`} to={to}>
                  <Icon aria-hidden="true" size={22} strokeWidth={1.7} />
                  <span>{label}</span>
                  <p>{description}</p>
                  <ArrowRight aria-hidden="true" className="tool-card-arrow" size={17} />
                </Link>
              ))}
            </div>
          </section>
        ))}
      </section>

      <section className="proof-callout" aria-labelledby="proof-title">
        <ShieldCheck aria-hidden="true" size={26} strokeWidth={1.6} />
        <div>
          <h2 id="proof-title">See the current application, not a glossy claim sheet.</h2>
          <p>Read the product map or browse fresh captures of the working screens.</p>
        </div>
        <div className="proof-actions">
          <Link className="button button-secondary" to="/comparison">Product map</Link>
          <Link className="button button-secondary" to="/screenshots">Screens</Link>
        </div>
      </section>
    </div>
  )
}

export default Home
