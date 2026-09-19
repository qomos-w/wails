package application

import (
	"sync"
)

// navCommandTracker correlates host navigation commands (Navigate / GoBack /
// GoForward / Reload) with the top-level NavigationStarting event they produce,
// per the browser-navigation contract §2.5.2.
//
// The adapter (platform layer) allocates a commandToken via beginCommand before
// invoking the COM navigation primitive, then hands that token to the first
// top-level, non-redirect NavigationStarting via consumeStarting. The host
// (sporecode) matches the token against its pending command and never inspects
// native NavigationId frontiers (§2.5.2 reliability rationale).
//
// The tracker holds only correlation state; it has no COM or platform
// dependencies and is safe for concurrent use.
type navCommandTracker struct {
	mu            sync.Mutex
	nextToken     uint64 // monotonic per-window counter, starting at 1
	expecting     uint64 // token awaiting its first-hop Starting; 0 = none
	expectingKind string // "navigate" | "back" | "forward" | "reload" (diagnostic)
}

// beginCommand allocates a fresh, monotonically increasing commandToken and
// records it as the single currently-expecting token. Returns the token. The
// caller must arrange for either consumeStarting (navigation happened) or
// clearExpecting (no navigation) to resolve it.
func (t *navCommandTracker) beginCommand(kind string) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextToken++
	t.expecting = t.nextToken
	t.expectingKind = kind
	return t.expecting
}

// consumeStarting hands the currently-expecting token to a first-hop
// (non-redirect) NavigationStarting and clears the expectation. Returns
// (token, true) when a command was expecting, or (0, false) when no host
// command was outstanding (page-initiated navigation, or a late Starting after a
// no-navigation/timeout resolution).
func (t *navCommandTracker) consumeStarting() (uint64, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.expecting == 0 {
		return 0, false
	}
	token := t.expecting
	t.expecting = 0
	t.expectingKind = ""
	return token, true
}

// clearExpecting discards the currently-expecting token without matching a
// Starting, returning it. Used by the no-navigation resolution paths (CanGoBack
// pre-check / timeout). Returns (0, false) when nothing was expecting.
func (t *navCommandTracker) clearExpecting() (uint64, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.expecting == 0 {
		return 0, false
	}
	token := t.expecting
	t.expecting = 0
	t.expectingKind = ""
	return token, true
}

// expectingToken returns the currently-expecting token (0 if none), for
// diagnostics. It must not be used for attribution decisions (see §2.5.2).
func (t *navCommandTracker) expectingToken() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.expecting
}
