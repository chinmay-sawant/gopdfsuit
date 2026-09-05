const OpPageShell = ({
  title,
  icon,
  description,
  children,
  steps = null,
}) => (
  <div className="tool-page">
    <header className="tool-page-header">
      <h1 className="tool-page-title">
        {icon}
        {title}
      </h1>
      {description && <p className="tool-page-description">{description}</p>}
    </header>
    <section className="tool-page-content">
        {children}
        {steps && steps.length > 0 && (
          <div className="tool-page-steps" aria-label="How to use this tool">
            {steps.map((step, index) => (
              <article className="tool-page-step" key={step.title}>
                <span className="tool-page-step-number">{index + 1}</span>
                <h3>{step.title}</h3>
                <p>{step.desc}</p>
              </article>
            ))}
          </div>
        )}
    </section>
  </div>
)

export default OpPageShell
