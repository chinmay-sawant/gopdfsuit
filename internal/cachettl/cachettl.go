// Package cachettl owns the shared time-to-live policy for process-lifetime
// content caches (font subsets, page compress output, font objects, prop
// parses, decoded images, signers, template-data files).
//
// Pools (sync.Pool buffers, zlib writers) are not affected: they recycle
// memory without reusing document content. Only content caches expire.
//
// Default is 3 minutes. Override with the GOPDFSUIT_CACHE_TTL environment
// variable (Go duration string, e.g. "2m", "90s") or SetCacheTTL in code.
// A non-positive TTL disables expiry and restores the previous
// size-only eviction behavior.
package cachettl

import (
	"os"
	"sync/atomic"
	"time"
)

// DefaultTTL is the out-of-box content cache lifetime.
const DefaultTTL = 3 * time.Minute

// EnvVar names the environment override read once at startup.
const EnvVar = "GOPDFSUIT_CACHE_TTL"

var ttlNanos atomic.Int64

func init() {
	ttlNanos.Store(int64(DefaultTTL))
	if raw := os.Getenv(EnvVar); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			ttlNanos.Store(int64(d))
		}
	}
}

// SetCacheTTL overrides the shared content cache lifetime. A value <= 0
// disables time-based expiry.
func SetCacheTTL(d time.Duration) {
	ttlNanos.Store(int64(d))
}

// Get returns the current content cache lifetime. Values <= 0 mean expiry
// is disabled.
func Get() time.Duration {
	return time.Duration(ttlNanos.Load())
}

// ExpiresAt returns the instant a newly stored entry stops being valid.
// Callers store it alongside the cached value.
func ExpiresAt(now time.Time) time.Time {
	ttl := Get()
	if ttl <= 0 {
		return time.Time{}
	}
	return now.Add(ttl)
}

// Expired reports whether an entry stored with the given expiry instant is
// no longer valid. The zero instant never expires (TTL disabled path).
func Expired(expiresAt time.Time, now time.Time) bool {
	if expiresAt.IsZero() {
		return false
	}
	return !now.Before(expiresAt)
}
