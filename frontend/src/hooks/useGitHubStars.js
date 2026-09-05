import { useEffect, useState } from 'react'

const REPO = 'chinmay-sawant/gopdfsuit'
const API_URL = `https://api.github.com/repos/${REPO}`
const STORAGE_KEY = 'gopdfsuit:github-stars'

// One network call per document at most. Client-side route changes reuse
// the cached value; only a full document reload refreshes it.
let fetchStarted = false

function readCache() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    return typeof parsed?.stars === 'number' ? parsed.stars : null
  } catch {
    return null
  }
}

function writeCache(stars) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ stars, updatedAt: Date.now() }))
  } catch {
    // Private mode or disabled storage: stay memory-only.
  }
}

function wasReloaded() {
  try {
    return performance.getEntriesByType('navigation')[0]?.type === 'reload'
  } catch {
    return false
  }
}

export function formatStars(count) {
  if (count >= 1000) return `${(count / 1000).toFixed(1).replace(/\.0$/, '')}k`
  return `${count}`
}

export function useGitHubStars() {
  const [stars, setStars] = useState(readCache)

  useEffect(() => {
    if (fetchStarted) return
    // Cached value wins on client-side navigation. Only a full document
    // reload (including hard refresh) refreshes the count.
    if (stars !== null && !wasReloaded()) return
    fetchStarted = true
    let cancelled = false
    fetch(API_URL, { cache: 'no-store' })
      .then((response) => {
        if (!response.ok) throw new Error(`GitHub API responded with ${response.status}`)
        return response.json()
      })
      .then((data) => {
        if (cancelled || typeof data?.stargazers_count !== 'number') return
        setStars(data.stargazers_count)
        writeCache(data.stargazers_count)
      })
      .catch(() => {
        // Offline or rate-limited: keep the cached value, or stay hidden
        // when there is none. Never render an invented count.
      })
    return () => {
      cancelled = true
    }
  }, [stars])

  return stars
}
