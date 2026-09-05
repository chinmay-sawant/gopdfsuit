import { formatStars, useGitHubStars } from '../../hooks/useGitHubStars'

export const GITHUB_REPO_URL = 'https://github.com/chinmay-sawant/gopdfsuit'

export default function GitHubStars({ className = '', showLabelWhileLoading = true }) {
  const stars = useGitHubStars()
  return (
    <a
      aria-label={stars === null ? 'Star GoPdfSuit on GitHub' : `Star GoPdfSuit on GitHub, currently ${stars} stars`}
      className={className}
      href={GITHUB_REPO_URL}
      rel="noreferrer"
      target="_blank"
    >
      <span aria-hidden="true">⭐</span>
      <span>{stars === null ? (showLabelWhileLoading ? 'Star' : '') : formatStars(stars)}</span>
    </a>
  )
}
