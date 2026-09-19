package application

// This file defines the navigation payload carried by window events for the
// WebView2 top-level navigation lifecycle (see browser-navigation contract
// §11.A). The adapter (platform layer) populates a WindowEventContext via the
// setters below and emits it through the windowEvents channel; the host
// (sporecode) reads it via the typed getters.
//
// NavigationStarting carries:
//   - NavigationID: the WebView2-native navigation identity (stable across the
//     navigation's redirect hops and its Completed event).
//   - URL: the navigation's requested URI.
//   - IsRedirected / IsUserInitiated: redirect-hop / user-gesture flags.
//   - CommandToken: the adapter host-command correlation token (§2.5.2). 0 means
//     the navigation was page-initiated (or unknown), non-zero correlates it to a
//     specific host Navigate/GoBack/GoForward/Reload command.
//   - NavStarted: false marks an explicit "no navigation" correlation answer
//     (§2.5.2 step 4) for a host command that produced no top-level navigation.
//
// NavigationCompleted carries:
//   - NavigationID, URL (post-redirect final URL via GetSource), IsSuccess,
//     WebErrorStatus.
const (
	navCtxKeyNavigationID    = "navigationId"
	navCtxKeyURL             = "url"
	navCtxKeyIsRedirected    = "isRedirected"
	navCtxKeyIsUserInitiated = "isUserInitiated"
	navCtxKeyCommandToken    = "commandToken"
	navCtxKeyIsSuccess       = "isSuccess"
	navCtxKeyWebErrorStatus  = "webErrorStatus"
	navCtxKeyNavStarted      = "navStarted"
)

// newWindowEventWithContext builds a WindowEvent that carries the given context,
// or an empty WindowEvent when ctx is nil (preserving legacy payload-less
// behaviour for events that carry no data).
func newWindowEventWithContext(ctx *WindowEventContext) *WindowEvent {
	ev := NewWindowEvent()
	if ctx == nil {
		// Legacy payload-less events still get an empty (non-nil) context so that
		// consumers reading navigation getters never dereference a nil Context().
		ctx = newWindowEventContext()
	}
	ev.ctx = ctx
	return ev
}

func (c *WindowEventContext) set(key string, value any) {
	if c.data == nil {
		c.data = make(map[string]any)
	}
	c.data[key] = value
}

// setNavigationStarting populates a NavigationStarting payload.
func (c *WindowEventContext) setNavigationStarting(navigationID uint64, url string, isRedirected, isUserInitiated bool, commandToken uint64) {
	c.set(navCtxKeyNavigationID, navigationID)
	c.set(navCtxKeyURL, url)
	c.set(navCtxKeyIsRedirected, isRedirected)
	c.set(navCtxKeyIsUserInitiated, isUserInitiated)
	c.set(navCtxKeyCommandToken, commandToken)
	c.set(navCtxKeyNavStarted, true)
}

// setNoNavigation marks an explicit no-navigation correlation answer for a host
// command that produced no top-level navigation (§2.5.2 step 4). Only the
// command token is meaningful.
func (c *WindowEventContext) setNoNavigation(commandToken uint64) {
	c.set(navCtxKeyCommandToken, commandToken)
	c.set(navCtxKeyNavStarted, false)
}

// setNavigationCompleted populates a NavigationCompleted payload with the
// post-redirect final URL.
func (c *WindowEventContext) setNavigationCompleted(navigationID uint64, url string, isSuccess bool, webErrorStatus int32) {
	c.set(navCtxKeyNavigationID, navigationID)
	c.set(navCtxKeyURL, url)
	c.set(navCtxKeyIsSuccess, isSuccess)
	c.set(navCtxKeyWebErrorStatus, webErrorStatus)
}

// --- getters (consumed by the host / sporecode) ---

func (c WindowEventContext) get(key string) any {
	if c.data == nil {
		return nil
	}
	return c.data[key]
}

// NavigationID returns the WebView2-native NavigationId of the event.
func (c WindowEventContext) NavigationID() uint64 {
	if v, ok := c.get(navCtxKeyNavigationID).(uint64); ok {
		return v
	}
	return 0
}

// URL returns the navigation URL: the requested URI for NavigationStarting, the
// post-redirect final URL for NavigationCompleted.
func (c WindowEventContext) URL() string {
	if v, ok := c.get(navCtxKeyURL).(string); ok {
		return v
	}
	return ""
}

// IsRedirected reports whether the NavigationStarting event is a redirect hop.
func (c WindowEventContext) IsRedirected() bool {
	if v, ok := c.get(navCtxKeyIsRedirected).(bool); ok {
		return v
	}
	return false
}

// IsUserInitiated reports whether the navigation was initiated by a user gesture.
func (c WindowEventContext) IsUserInitiated() bool {
	if v, ok := c.get(navCtxKeyIsUserInitiated).(bool); ok {
		return v
	}
	return false
}

// CommandToken returns the adapter host-command correlation token. 0 means the
// navigation was page-initiated or the token is not applicable.
func (c WindowEventContext) CommandToken() uint64 {
	if v, ok := c.get(navCtxKeyCommandToken).(uint64); ok {
		return v
	}
	return 0
}

// IsSuccess reports whether a NavigationCompleted event succeeded.
func (c WindowEventContext) IsSuccess() bool {
	if v, ok := c.get(navCtxKeyIsSuccess).(bool); ok {
		return v
	}
	return false
}

// WebErrorStatus returns the COREWEBVIEW2_WEB_ERROR_STATUS of a failed
// NavigationCompleted (meaningful only when IsSuccess is false).
func (c WindowEventContext) WebErrorStatus() int32 {
	if v, ok := c.get(navCtxKeyWebErrorStatus).(int32); ok {
		return v
	}
	return 0
}

// NavStarted reports whether the navigation actually started. For a real
// NavigationStarting event it is true; for an explicit no-navigation
// correlation answer it is false.
func (c WindowEventContext) NavStarted() bool {
	if v, ok := c.get(navCtxKeyNavStarted).(bool); ok {
		return v
	}
	return true
}
