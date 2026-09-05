import { ArrowRight, Braces, FileOutput, ShieldCheck, Wrench } from 'lucide-react'
import { Link } from 'react-router-dom'

const workflows = [
  {
    title: 'Start with a template',
    icon: Braces,
    copy: 'Use the editor when you need to compose a template. Use the viewer when you already have JSON and need a PDF.',
    links: [
      { to: '/editor', label: 'Open the editor' },
      { to: '/viewer', label: 'Preview JSON' },
    ],
  },
  {
    title: 'Work with an existing PDF',
    icon: Wrench,
    copy: 'Merge, split, compress, fill forms, or redact. Each tool starts with the file and keeps the next action obvious.',
    links: [
      { to: '/merge', label: 'Merge PDFs' },
      { to: '/split', label: 'Split a PDF' },
      { to: '/compress', label: 'Compress a PDF' },
      { to: '/filler', label: 'Fill a form' },
      { to: '/redact', label: 'Redact content' },
    ],
  },
  {
    title: 'Start with HTML',
    icon: FileOutput,
    copy: 'Choose a document or image output, then use the matching HTML conversion workspace.',
    links: [
      { to: '/htmltopdf', label: 'HTML to PDF' },
      { to: '/htmltoimage', label: 'HTML to image' },
    ],
  },
]

export default function Comparison() {
  return (
    <div className="proof-page">
      <header>
        <h1>One toolkit. Clear paths through a document job.</h1>
        <p>This page maps the jobs GoPdfSuit covers. It does not make pricing, benchmark, or competitor claims without a current source.</p>
      </header>

      <section className="workflow-map" aria-label="GoPdfSuit workflows">
        {workflows.map(({ title, copy, icon: Icon, links }) => (
          <article className="workflow-card" key={title}>
            <Icon aria-hidden="true" size={24} strokeWidth={1.6} />
            <h2>{title}</h2>
            <p>{copy}</p>
            <ul>
              {links.map(({ to, label }) => <li key={to}><Link to={to}>{label}</Link></li>)}
            </ul>
          </article>
        ))}
        <article className="workflow-card">
          <ShieldCheck aria-hidden="true" size={24} strokeWidth={1.6} />
          <h2>Know where the work runs.</h2>
          <p>Tools describe their browser and service behavior near the action, so a file transfer is never hidden in marketing copy.</p>
          <Link to="/screenshots">Browse current screens <ArrowRight aria-hidden="true" size={16} /></Link>
        </article>
      </section>
    </div>
  )
}
