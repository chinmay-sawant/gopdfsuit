import { Link } from 'react-router-dom'
import { Transition } from '@headlessui/react'
import {
  FileText,
  Edit,
  Merge,
  FileCheck,
  Globe,
  Image,
  Zap,
  Eraser,
  Sigma,
  ArrowRight,
} from 'lucide-react'

const features = [
  {
    icon: <FileText size={24} />,
    title: 'Native Go Support',
    description: 'Use as a standalone library (gopdflib) or via HTTP API.',
    link: 'https://github.com/chinmay-sawant/gopdfsuit/tree/master/pkg/gopdflib',
    color: 'blue',
    external: true,
  },
  {
    icon: <Globe size={24} />,
    title: 'Python Web Client',
    description: 'Lightweight API client for interacting with the GoPdfSuit server.',
    link: 'https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/python/gopdf',
    color: 'teal',
    external: true,
  },
  {
    icon: <Zap size={24} />,
    title: 'Native Python Support',
    description: 'High-performance CGO bindings for direct PDF generation from Python.',
    link: 'https://github.com/chinmay-sawant/gopdfsuit/tree/master/bindings/python',
    color: 'yellow',
    external: true,
  },
  {
    icon: <Globe size={24} />,
    title: 'Language Agnostic',
    description: 'REST API works with any programming language.',
    link: '#section-api',
    color: 'purple',
    external: false,
  },
  {
    icon: <FileText size={24} />,
    title: 'Template-based PDF',
    description: 'JSON-driven PDF creation with multi-page support and automatic page breaks.',
    link: '/viewer',
    color: 'teal',
  },
  {
    icon: <Edit size={24} />,
    title: 'Visual PDF Editor',
    description: 'Drag-and-drop interface for building PDF templates with live preview.',
    link: '/editor',
    color: 'blue',
  },
  {
    icon: <Merge size={24} />,
    title: 'PDF Merge',
    description: 'Combine multiple PDFs with drag-and-drop reordering and live preview.',
    link: '/merge',
    color: 'purple',
  },
  {
    icon: <FileCheck size={24} />,
    title: 'Form Filling',
    description: 'AcroForm and XFDF support for filling PDF forms programmatically.',
    link: '/filler',
    color: 'yellow',
  },
  {
    icon: <Globe size={24} />,
    title: 'HTML to PDF',
    description: 'Convert HTML content or web pages to PDF using Chromium.',
    link: '/htmltopdf',
    color: 'green',
  },
  {
    icon: <Image size={24} />,
    title: 'HTML to Image',
    description: 'Convert HTML to PNG, JPG, or SVG with custom dimensions.',
    link: '/htmltoimage',
    color: 'blue',
  },
  {
    icon: <Eraser size={24} />,
    title: 'PDF Redaction',
    description: 'Redact sensitive content with visual selection or text search across pages.',
    link: '/redact',
    color: 'red',
  },
  {
    icon: <Sigma size={24} />,
    title: 'Typst Math Rendering',
    description: 'Render mathematical equations in PDFs using Typst syntax with full symbol support.',
    link: 'https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/typstsyntax',
    color: 'green',
    external: true,
  },
]

const FeatureCard = ({ feature }) => (
  <div className={`glass-card feature-card-inner`}>
    <div className="feature-card-content">
      <div className="feature-card-header">
        <div
          className={`feature-icon-box ${feature.color}`}
          style={{ width: '48px', height: '48px' }}
        >
          {feature.icon}
        </div>
        <h3>{feature.title}</h3>
      </div>
      <p className="feature-card-desc">{feature.description}</p>
      <div className="feature-card-link">
        {feature.external ? 'View on GitHub' : 'Try it now'}
        <ArrowRight size={14} />
      </div>
    </div>
    {feature.wide && feature.extra && (
      <div className="feature-card-extra">
        {feature.extra}
      </div>
    )}
  </div>
)

const FeaturesSection = ({ isVisible }) => {
  const visible = isVisible['section-features']

  return (
    <section id="section-features" style={{ padding: '5rem 0' }}>
      <div className="container">
        <Transition
          as="div"
          show={!!visible}
          enter="animate-fadeInUp"
          enterFrom="stagger-animation"
          enterTo="stagger-animation visible"
          className="text-center"
          style={{ marginBottom: '3rem' }}
        >
          <h2 className="gradient-text section-heading">
            Powerful Features
          </h2>
          <p className="section-subheading">
            Everything you need for professional PDF workflows
          </p>
        </Transition>

        <div className="bento-features">
          {features.map((feature, index) => {
            const card = (
              <Transition
                as="div"
                show={!!visible}
                enter="animate-fadeInScale"
                enterFrom="stagger-animation"
                enterTo="stagger-animation visible"
                style={{ animationDelay: `${0.1 + index * 0.06}s` }}
              >
                <FeatureCard feature={feature} />
              </Transition>
            )

            const wrapperClass = feature.wide ? 'bento-wide' : ''

            if (feature.link.startsWith('#')) {
              return (
                <div
                  key={index}
                  className={wrapperClass}
                  onClick={() => {
                    const element = document.getElementById(feature.link.substring(1))
                    if (element) element.scrollIntoView({ behavior: 'smooth' })
                  }}
                  style={{ textDecoration: 'none', color: 'inherit', cursor: 'pointer' }}
                >
                  {card}
                </div>
              )
            }

            return feature.external ? (
              <a
                key={index}
                className={wrapperClass}
                href={feature.link}
                target="_blank"
                rel="noopener noreferrer"
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                {card}
              </a>
            ) : (
              <Link
                key={index}
                className={wrapperClass}
                to={feature.link}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                {card}
              </Link>
            )
          })}
        </div>
      </div>
    </section>
  )
}

export default FeaturesSection
