// Package edge re-exports the WebView2 edge types that live in the wails v3
// internal tree. The internal package cannot be imported outside the wails
// module, but applications still need to name cookie-manager types returned by
// the public application API (e.g. WebviewWindow.GetCookieManager).
package edge

import (
	edgeinternal "github.com/wailsapp/wails/v3/internal/webview2/pkg/edge"
)

type ICoreWebView2CookieManager = edgeinternal.ICoreWebView2CookieManager

type ICoreWebView2CookieList = edgeinternal.ICoreWebView2CookieList

type ICoreWebView2Cookie = edgeinternal.ICoreWebView2Cookie

type GetCookiesCompletion = edgeinternal.GetCookiesCompletion
