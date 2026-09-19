//go:build windows

package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// _ICoreWebView2NavigationStartingEventArgsVtbl mirrors the COM vtable of
// ICoreWebView2NavigationStartingEventArgs. Slot order follows the WebView2 IDL:
// GetUri, GetIsUserInitiated, GetIsRedirected, GetNavigationId, GetRequestHeaders,
// GetCancel, PutCancel.
type _ICoreWebView2NavigationStartingEventArgsVtbl struct {
	_IUnknownVtbl
	GetUri             ComProc
	GetIsUserInitiated ComProc
	GetIsRedirected    ComProc
	GetNavigationId    ComProc
	GetRequestHeaders  ComProc
	GetCancel          ComProc
	PutCancel          ComProc
}

type ICoreWebView2NavigationStartingEventArgs struct {
	vtbl *_ICoreWebView2NavigationStartingEventArgsVtbl
}

func (i *ICoreWebView2NavigationStartingEventArgs) AddRef() uint32 {
	ret, _, _ := i.vtbl.AddRef.Call(uintptr(unsafe.Pointer(i)))
	return uint32(ret)
}

func (i *ICoreWebView2NavigationStartingEventArgs) Release() uint32 {
	ret, _, _ := i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))
	return uint32(ret)
}

// GetUri returns the URI of the navigation that is starting.
func (i *ICoreWebView2NavigationStartingEventArgs) GetUri() (string, error) {
	var _uri *uint16
	hr, _, _ := i.vtbl.GetUri.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&_uri)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return "", windows.Errno(hr)
	}
	uri := windows.UTF16PtrToString(_uri)
	windows.CoTaskMemFree(unsafe.Pointer(_uri))
	return uri, nil
}

// GetIsUserInitiated returns whether the navigation was initiated by a user
// gesture (as opposed to script/redirect).
func (i *ICoreWebView2NavigationStartingEventArgs) GetIsUserInitiated() (bool, error) {
	var result bool
	hr, _, _ := i.vtbl.GetIsUserInitiated.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&result)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return false, windows.Errno(hr)
	}
	return result, nil
}

// GetIsRedirected returns whether the navigation is a redirect (server-side,
// meta refresh, or same-hop jump). Redirect hops share the originating
// navigation's NavigationId.
func (i *ICoreWebView2NavigationStartingEventArgs) GetIsRedirected() (bool, error) {
	var result bool
	hr, _, _ := i.vtbl.GetIsRedirected.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&result)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return false, windows.Errno(hr)
	}
	return result, nil
}

// GetNavigationId returns the WebView2-native NavigationId. It is consistent
// across the Starting and Completed events of a single top-level navigation
// (including redirect hops) and unique per navigation within a webview.
func (i *ICoreWebView2NavigationStartingEventArgs) GetNavigationId() (uint64, error) {
	var navigationId uint64
	hr, _, _ := i.vtbl.GetNavigationId.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&navigationId)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, windows.Errno(hr)
	}
	return navigationId, nil
}

// PutCancel cancels the in-progress navigation.
func (i *ICoreWebView2NavigationStartingEventArgs) PutCancel(cancel bool) error {
	hr, _, _ := i.vtbl.PutCancel.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(boolToInt(cancel)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return windows.Errno(hr)
	}
	return nil
}
