package application

import "testing"

// TestNavigationStartingContextRoundTrip verifies the adapter-populated
// NavigationStarting payload is readable via the typed getters the host
// (sporecode) consumes (§11.A / §2.5).
func TestNavigationStartingContextRoundTrip(t *testing.T) {
	ctx := newWindowEventContext()
	ctx.setNavigationStarting(42, "https://example.com/page", false, true, 7)

	if ctx.NavigationID() != 42 {
		t.Errorf("NavigationID = %d, want 42", ctx.NavigationID())
	}
	if ctx.URL() != "https://example.com/page" {
		t.Errorf("URL = %q, want https://example.com/page", ctx.URL())
	}
	if ctx.IsRedirected() {
		t.Error("IsRedirected = true, want false")
	}
	if !ctx.IsUserInitiated() {
		t.Error("IsUserInitiated = false, want true")
	}
	if ctx.CommandToken() != 7 {
		t.Errorf("CommandToken = %d, want 7", ctx.CommandToken())
	}
	if !ctx.NavStarted() {
		t.Error("NavStarted = false, want true for a real Starting")
	}
}

// TestNavigationStartingRedirectHop verifies a redirect hop carries no host
// command token (the adapter does not consume the expecting token for redirects,
// §2.5.2 step 3) and shares the originating navigation's identity.
func TestNavigationStartingRedirectHop(t *testing.T) {
	ctx := newWindowEventContext()
	ctx.setNavigationStarting(42, "https://example.com/redirected", true, false, 0)

	if ctx.NavigationID() != 42 {
		t.Errorf("redirect must share NavigationId 42, got %d", ctx.NavigationID())
	}
	if !ctx.IsRedirected() {
		t.Error("IsRedirected = false, want true")
	}
	if ctx.CommandToken() != 0 {
		t.Errorf("redirect hop must carry commandToken 0, got %d", ctx.CommandToken())
	}
}

// TestNoNavigationContext verifies the explicit no-navigation correlation answer
// (§2.5.2 step 4) carries the token with NavStarted=false.
func TestNoNavigationContext(t *testing.T) {
	ctx := newWindowEventContext()
	ctx.setNoNavigation(9)

	if ctx.CommandToken() != 9 {
		t.Errorf("CommandToken = %d, want 9", ctx.CommandToken())
	}
	if ctx.NavStarted() {
		t.Error("NavStarted = true, want false for no-navigation answer")
	}
	if ctx.NavigationID() != 0 {
		t.Errorf("NavigationID = %d, want 0 for no-navigation answer", ctx.NavigationID())
	}
	if ctx.URL() != "" {
		t.Errorf("URL = %q, want empty for no-navigation answer", ctx.URL())
	}
}

// TestNavigationCompletedContextRoundTrip verifies the commit-boundary payload
// (§3): NavigationId, post-redirect final URL, IsSuccess, WebErrorStatus.
func TestNavigationCompletedContextRoundTrip(t *testing.T) {
	ctx := newWindowEventContext()
	ctx.setNavigationCompleted(42, "https://example.com/final", true, 0)

	if ctx.NavigationID() != 42 {
		t.Errorf("NavigationID = %d, want 42", ctx.NavigationID())
	}
	if ctx.URL() != "https://example.com/final" {
		t.Errorf("URL = %q, want https://example.com/final", ctx.URL())
	}
	if !ctx.IsSuccess() {
		t.Error("IsSuccess = false, want true")
	}
	if ctx.WebErrorStatus() != 0 {
		t.Errorf("WebErrorStatus = %d, want 0", ctx.WebErrorStatus())
	}
}

// TestNavigationCompletedFailed verifies a failed navigation surfaces its error
// status.
func TestNavigationCompletedFailed(t *testing.T) {
	ctx := newWindowEventContext()
	ctx.setNavigationCompleted(42, "", false, 7)

	if ctx.IsSuccess() {
		t.Error("IsSuccess = true, want false")
	}
	if ctx.WebErrorStatus() != 7 {
		t.Errorf("WebErrorStatus = %d, want 7", ctx.WebErrorStatus())
	}
}

// TestEmptyContextDefaults verifies that a payload-less event (legacy behaviour)
// yields safe zero-value defaults and NavStarted==true (so existing consumers
// that ignore the context are unaffected).
func TestEmptyContextDefaults(t *testing.T) {
	ev := newWindowEventWithContext(nil)
	if ev.Context() == nil {
		t.Fatal("legacy event context must not be nil (Context() must be safe to call)")
	}
	c := *ev.Context()
	if c.NavigationID() != 0 || c.URL() != "" || c.CommandToken() != 0 {
		t.Error("empty context getters must be zero-valued")
	}
	if !c.NavStarted() {
		t.Error("NavStarted must default to true so payload-less events are unaffected")
	}
}

// TestWindowEventWithContext verifies the context threads through the WindowEvent
// used by listeners.
func TestWindowEventWithContext(t *testing.T) {
	ctx := newWindowEventContext()
	ctx.setNavigationStarting(1, "https://example.com", false, false, 3)
	ev := newWindowEventWithContext(ctx)

	got := ev.Context()
	if got == nil {
		t.Fatal("Context() must return the threaded context")
	}
	if got.CommandToken() != 3 {
		t.Errorf("threaded CommandToken = %d, want 3", got.CommandToken())
	}
}
