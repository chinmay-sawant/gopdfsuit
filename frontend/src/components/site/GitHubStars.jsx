import { Star } from 'lucide-react'
import { formatStars, useGitHubStars } from '../../hooks/useGitHubStars'

export const GITHUB_REPO_URL = 'https://github.com/chinmay-sawant/gopdfsuit'

export default function GitHubStars({ className = '', iconSize = 14 }) {
  const stars = useGitHubStars()
  if (stars === null) return null
  return (
    <a
      aria-label={`Star GoPdfSuit on GitHub, currently ${stars} stars`}
      className={className}
      href={GITHUB_REPO_URL}
      rel="noreferrer"
      target="_blank"
    >
      <Star aria-hidden="true" fill="currentColor" size={iconSize} />
      <span>{formatStars(stars)}</span>
    </a>
  )
}
