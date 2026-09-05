import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <section className="not-found-page" aria-labelledby="not-found-title">
      <h1 id="not-found-title">That route does not exist.</h1>
      <p>Return to the tool catalogue and choose the PDF job you need.</p>
      <Link className="button button-primary" to="/">Go to the home page</Link>
    </section>
  )
}
