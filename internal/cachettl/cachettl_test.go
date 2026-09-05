package cachettl

import (
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	if got := Get(); got != DefaultTTL {
		t.Fatalf("default TTL = %v, want %v", got, DefaultTTL)
	}
}

func TestExpiryRoundTrip(t *testing.T) {
	old := Get()
	defer SetCacheTTL(old)

	SetCacheTTL(50 * time.Millisecond)
	now := time.Now()
	exp := ExpiresAt(now)
	if Expired(exp, now) {
		t.Fatal("fresh entry must not be expired")
	}
	if !Expired(exp, now.Add(100*time.Millisecond)) {
		t.Fatal("entry past TTL must be expired")
	}
}

func TestDisabledTTLNeverExpires(t *testing.T) {
	old := Get()
	defer SetCacheTTL(old)

	SetCacheTTL(0)
	exp := ExpiresAt(time.Now())
	if !exp.IsZero() {
		t.Fatal("disabled TTL must store zero instant")
	}
	if Expired(exp, time.Now().Add(time.Hour)) {
		t.Fatal("disabled TTL must never expire")
	}
}
