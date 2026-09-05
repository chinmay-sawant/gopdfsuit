import { showcaseCaptures } from '../content/showcase'

export default function Screenshots() {
  const base = import.meta.env.BASE_URL

  return (
    <div className="screens-page">
      <header>
        <h1>The application as it is built.</h1>
        <p>These local captures ship with the site. They do not depend on a mutable branch or a remote image host.</p>
      </header>

      <section className="screen-grid" aria-label="GoPdfSuit application screenshots">
        {showcaseCaptures.map(({ file, title, description, viewport }) => (
          <article className="screen-card" key={file}>
            <figure>
              <img alt={`${title}. ${description}`} src={`${base}showcase/${file}`} />
              <figcaption>
                <h2>{title}</h2>
                <p>{description}</p>
                <small>Captured 5 September 2026 at {viewport}.</small>
              </figcaption>
            </figure>
          </article>
        ))}
      </section>
    </div>
  )
}
