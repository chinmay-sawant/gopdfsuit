import BackgroundAnimation from './BackgroundAnimation'

/**
 * OpPageShell owns the shared op-page chrome every tool page pasted by hand:
 * background animation, centered hero header (badge, title, description),
 * and the optional How to Use card. Page-specific form + result grid is
 * passed as children so layout stays per-op.
 */
const OpPageShell = ({
  badge,
  badgeTone = 'rgba(78,205,196,0.1)',
  badgeBorder = 'rgba(78,205,196,0.3)',
  badgeColor = '#4ecdc4',
  title,
  icon,
  description,
  children,
  steps = null,
  stepsTitle = 'How to Use',
}) => (
  <div style={{ minHeight: '100vh', position: 'relative' }}>
    <BackgroundAnimation />
    <section style={{ padding: '4rem 0 2rem', textAlign: 'center' }}>
      <div className="container">
        {badge && (
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem', background: badgeTone, border: `1px solid ${badgeBorder}`, borderRadius: '50px', marginBottom: '1.5rem', color: badgeColor, fontSize: '0.9rem', fontWeight: '500' }}>
            {badge}
          </div>
        )}
        <h1 style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem', marginBottom: '1rem', fontSize: 'clamp(2rem,5vw,3rem)', fontWeight: '800', color: 'hsl(var(--foreground))' }}>
          {icon}
          {title}
        </h1>
        {description && (
          <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '1.1rem', maxWidth: '600px', margin: '0 auto' }}>{description}</p>
        )}
      </div>
    </section>

    <section style={{ padding: '2rem 0 4rem' }}>
      <div className="container">
        {children}
        {steps && steps.length > 0 && (
          <div className="glass-card" style={{ marginTop: '2rem', padding: '2rem' }}>
            <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '1.1rem', fontWeight: '700' }}>
              <div className="feature-icon-box yellow" style={{ width: '40px', height: '40px', marginBottom: 0 }}><span style={{ fontSize: '1.2rem' }}>📋</span></div>{stepsTitle}
            </h3>
            <div className="grid grid-3" style={{ gap: '1.5rem' }}>
              {steps.map((step, i) => (
                <div key={i} style={{ textAlign: 'center', padding: '1rem', background: 'rgba(255,255,255,0.02)', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.05)' }}>
                  <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>{step.num}</div>
                  <h4 style={{ color: '#4ecdc4', marginBottom: '0.5rem', fontSize: '1rem' }}>{step.title}</h4>
                  <p style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.85rem', marginBottom: 0 }}>{step.desc}</p>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </section>
  </div>
)

export default OpPageShell
