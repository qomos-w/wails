//go:build windows

package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type _ICoreWebView2NavigationCompletedEventArgsVtbl struct {
	_IUnknownVtbl
	GetIsSuccess      ComProc
	GetWebErrorStatus ComProc
	GetNavigationId   ComProc
}

type ICoreWebView2NavigationCompletedEventArgs struct {
	vtbl *_ICoreWebView2NavigationCompletedEventArgsVtbl
}

func (i *ICoreWebView2NavigationCompletedEventArgs) AddRef() uint32 {
	ret, _, _ := i.vtbl.AddRef.Call(uintptr(unsafe.Pointer(i)))

	return uint32(ret)
}

func (i *ICoreWebView2NavigationCompletedEventArgs) Release() uint32 {
	ret, _, _ := i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))

	return uint32(ret)
}

// GetIsSuccess returns whether the navigation completed successfully.
func (i *ICoreWebView2NavigationCompletedEventArgs) GetIsSuccess() (bool, error) {
	var result bool
	hr, _, _ := i.vtbl.GetIsSuccess.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&result)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return false, windows.Errno(hr)
	}
	return result, nil
}

// GetWebErrorStatus returns the COREWEBVIEW2_WEB_ERROR_STATUS value if the
// navigation failed. It is only meaningful when GetIsSuccess returns false.
func (i *ICoreWebView2NavigationCompletedEventArgs) GetWebErrorStatus() (int32, error) {
	var status int32
	hr, _, _ := i.vtbl.GetWebErrorStatus.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&status)),
	)
	if windows.Handle(hr) != windows.S_OK {
		return 0, windows.Errno(hr)
	}
	return status, nil
}

// GetNavigationId returns the WebView2-native NavigationId of the completed
// navigation. It matches the NavigationId reported by the corresponding
// NavigationStarting event (including redirect hops).
func (i *ICoreWebView2NavigationCompletedEventArgs) GetNavigationId() (uint64, error) {
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
