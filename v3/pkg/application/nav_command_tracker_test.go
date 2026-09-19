package application

import (
	"testing"
)

// TestNavCommandTrackerMonotonicTokens verifies beginCommand allocates
// monotonically increasing tokens starting at 1 (§2.5.2 step 1).
func TestNavCommandTrackerMonotonicTokens(t *testing.T) {
	var tr navCommandTracker

	t1 := tr.beginCommand("navigate")
	t2 := tr.beginCommand("back")
	t3 := tr.beginCommand("reload")
	if t1 != 1 || t2 != 2 || t3 != 3 {
		t.Fatalf("tokens must be monotonic from 1, got %d,%d,%d", t1, t2, t3)
	}
}

// TestNavCommandTrackerConsumeStarting verifies a first-hop NavigationStarting
// hands over the currently expecting token and clears the expectation (§2.5.2.2).
func TestNavCommandTrackerConsumeStarting(t *testing.T) {
	var tr navCommandTracker

	token := tr.beginCommand("navigate")
	got, ok := tr.consumeStarting()
	if !ok || got != token {
		t.Fatalf("consumeStarting must hand over the expecting token %d, got %d (ok=%v)", token, got, ok)
	}

	// A second consumeStarting when nothing is expecting must report none.
	got, ok = tr.consumeStarting()
	if ok || got != 0 {
		t.Fatalf("consumeStarting with no command outstanding must return (0,false), got (%d,%v)", got, ok)
	}
}

// TestNavCommandTrackerPageInitiatedNavigation verifies that a page-initiated
// navigation (no host command outstanding) yields commandToken==0 (§2.5.4).
func TestNavCommandTrackerPageInitiatedNavigation(t *testing.T) {
	var tr navCommandTracker

	got, ok := tr.consumeStarting()
	if ok || got != 0 {
		t.Fatalf("page-initiated navigation must yield (0,false), got (%d,%v)", got, ok)
	}
}

// TestNavCommandTrackerClearExpecting verifies the no-navigation resolution
// (§2.5.2 step 4) discards the expecting token and makes a subsequent
// NavigationStarting report page-initiated.
func TestNavCommandTrackerClearExpecting(t *testing.T) {
	var tr navCommandTracker

	token := tr.beginCommand("forward")
	cleared, ok := tr.clearExpecting()
	if !ok || cleared != token {
		t.Fatalf("clearExpecting must return the token %d, got %d (ok=%v)", token, cleared, ok)
	}

	// After clearing, a late NavigationStarting must not carry the stale token.
	got, ok := tr.consumeStarting()
	if ok || got != 0 {
		t.Fatalf("after clearExpecting a Starting must yield (0,false), got (%d,%v)", got, ok)
	}
}

// TestNavCommandTrackerClearExpectingIdle verifies clearExpecting on an idle
// tracker reports nothing.
func TestNavCommandTrackerClearExpectingIdle(t *testing.T) {
	var tr navCommandTracker

	cleared, ok := tr.clearExpecting()
	if ok || cleared != 0 {
		t.Fatalf("clearExpecting when idle must return (0,false), got (%d,%v)", cleared, ok)
	}
}

// TestNavCommandTrackerSerialExpectation verifies the per-window invariant that
// at most one token is outstanding at a time (§2.5.1): a new beginCommand
// supersedes any unresolved expectation, and only the latest is consumable.
func TestNavCommandTrackerSerialExpectation(t *testing.T) {
	var tr navCommandTracker

	first := tr.beginCommand("navigate")
	second := tr.beginCommand("reload")
	if first == second {
		t.Fatal("two commands must yield distinct tokens")
	}

	got, ok := tr.consumeStarting()
	if !ok || got != second {
		t.Fatalf("only the latest command's token must be consumable, got %d want %d", got, second)
	}
	if got == first {
		t.Fatal("the superseded first token must never be handed over")
	}
}

// TestNavCommandTrackerExpectingToken verifies the diagnostic accessor.
func TestNavCommandTrackerExpectingToken(t *testing.T) {
	var tr navCommandTracker

	if tr.expectingToken() != 0 {
		t.Fatal("idle tracker must report expecting token 0")
	}
	token := tr.beginCommand("navigate")
	if tr.expectingToken() != token {
		t.Fatalf("expectingToken must report %d, got %d", token, tr.expectingToken())
	}
	tr.consumeStarting()
	if tr.expectingToken() != 0 {
		t.Fatal("expectingToken must be 0 after consumeStarting")
	}
}
