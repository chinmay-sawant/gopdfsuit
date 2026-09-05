import { useEffect, useState } from 'react'

const REPO = 'chinmay-sawant/gopdfsuit'
const API_URL = `https://api.github.com/repos/${REPO}`
const STORAGE_KEY = 'gopdfsuit:github-stars'

// One shared request per document. A shared promise (not a latch flag) is
// used so React StrictMode's setup-cleanup-setup pass in dev cannot orphan
// the in-flight request: every mount attaches to the same promise.
let inflight = null

function requestStars() {
  if (!inflight) {
    inflight = fetch(API_URL, { cache: 'no-store' })
      .then((response) => {
        if (!response.ok) throw new Error(`GitHub API responded with ${response.status}`)
        return response.json()
      })
      .then((data) => {
        if (typeof data?.stargazers_count !== 'number') throw new Error('Unexpected GitHub API response')
        return data.stargazers_count
      })
      .catch((error) => {
        inflight = null
        throw error
      })
  }
  return inflight
}

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
    // Cached value wins on client-side navigation. Only a full document
    // reload (including hard refresh) refreshes the count.
    if (stars !== null && !wasReloaded()) return
    let cancelled = false
    requestStars()
      .then((count) => {
        if (cancelled) return
        setStars(count)
        writeCache(count)
      })
      .catch(() => {
        // Offline or rate-limited: keep the cached value, or stay on the
        // "Star" label when there is none. Never render an invented count.
      })
    return () => {
      cancelled = true
    }
  }, [stars])

  return stars
}
