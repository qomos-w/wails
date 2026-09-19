//go:build windows

package edge

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ICoreWebView2_13 interface ID.
var iidCoreWebView2_13 = NewGUID("{f75f09a8-667e-4983-88d6-c8773f315e84}")

// Vtable slot (0-based) of ICoreWebView2_13::GetProfile.
// IUnknown(3) + ICoreWebView2(58) + _2(7) + _3(5) + _4(4) + _5(2) + _6(1) +
// _7(1) + _8(7) + _9(9) + _10(2) + _11(3) + _12(3) = 105.
const vtblGetProfile = 105

// Vtable slot (0-based) of ICoreWebView2Profile::PutPreferredColorScheme.
// IUnknown(3) + ProfileName + IsInPrivate + ProfilePath + GetDownload + PutDownload + GetColorScheme = 9.
const vtblPutPreferredColorScheme = 9

const (
	colorSchemeLight uint32 = 1
	colorSchemeDark  uint32 = 2
)

// SetPreferredColorScheme configures the WebView2 preferred color scheme so the
// prefers-color-scheme CSS media query reports the chosen mode to every page.
func (e *Chromium) SetPreferredColorScheme(dark bool) error {
	if e.webview == nil {
		return errors.New("webview not ready")
	}

	var wv13 unsafe.Pointer
	hr, _, _ := e.webview.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(iidCoreWebView2_13)),
		uintptr(unsafe.Pointer(&wv13)),
	)
	if windows.Handle(hr) != windows.S_OK || wv13 == nil {
		return errors.New("ICoreWebView2_13 unavailable")
	}
	defer comRelease(wv13)

	// GetProfile (slot 105) → *ICoreWebView2Profile.
	vtbl := *(*unsafe.Pointer)(wv13)
	var profile unsafe.Pointer
	hr, _, _ = (*[256]ComProc)(vtbl)[vtblGetProfile].Call(
		uintptr(wv13),
		uintptr(unsafe.Pointer(&profile)),
	)
	if windows.Handle(hr) != windows.S_OK || profile == nil {
		return errors.New("GetProfile failed")
	}
	defer comRelease(profile)

	// PutPreferredColorScheme (slot 9).
	pvtbl := *(*unsafe.Pointer)(profile)
	scheme := uintptr(colorSchemeLight)
	if dark {
		scheme = uintptr(colorSchemeDark)
	}
	hr, _, _ = (*[256]ComProc)(pvtbl)[vtblPutPreferredColorScheme].Call(
		uintptr(profile),
		scheme,
	)
	if windows.Handle(hr) != windows.S_OK {
		return errors.New("PutPreferredColorScheme failed")
	}
	return nil
}

// comRelease calls IUnknown::Release (vtable slot 2) on a raw COM interface.
func comRelease(p unsafe.Pointer) {
	if p == nil {
		return
	}
	vtbl := *(*unsafe.Pointer)(p)
	(*[3]ComProc)(vtbl)[2].Call(uintptr(p))
}
