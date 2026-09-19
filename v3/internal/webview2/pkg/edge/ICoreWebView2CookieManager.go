//go:build windows

package edge

import (
	"errors"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ICoreWebView2CookieManager vtable
type iCoreWebView2CookieManagerVtbl struct {
	_IUnknownVtbl
	CreateCookie                   ComProc
	CopyCookie                     ComProc
	GetCookies                     ComProc
	AddOrUpdateCookie              ComProc
	DeleteCookie                   ComProc
	DeleteCookies                  ComProc
	DeleteCookiesWithDomainAndPath ComProc
	DeleteAllCookies               ComProc
}

// ICoreWebView2CookieManager represents the cookie manager interface.
//
// This struct overlays a NATIVE COM object — only the vtbl field is safe to
// access. Do NOT add Go fields (mutex, waiter, etc.) here: they would write
// into the native object's internal memory and corrupt its state. Package-
// level maps keyed by the native pointer are used to track in-flight waiters
// instead, mirroring how Chromium keeps Go state off the ICoreWebView2 struct.
type ICoreWebView2CookieManager struct {
	vtbl *iCoreWebView2CookieManagerVtbl
}

// AddRef increments the reference count of ICoreWebView2CookieManager interface
func (i *ICoreWebView2CookieManager) AddRef() uint32 {
	ret, _, _ := i.vtbl.AddRef.Call(uintptr(unsafe.Pointer(i)))

	return uint32(ret)
}

// Release decrements the reference count of ICoreWebView2CookieManager interface
func (i *ICoreWebView2CookieManager) Release() uint32 {
	ret, _, _ := i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))

	return uint32(ret)
}

// CreateCookie creates a new cookie with the given parameters
func (i *ICoreWebView2CookieManager) CreateCookie(name, value, domain, path string) (*ICoreWebView2Cookie, error) {
	var cookie *ICoreWebView2Cookie

	nameutf16, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	valueutf16, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return nil, err
	}
	domainutf16, err := windows.UTF16PtrFromString(domain)
	if err != nil {
		return nil, err
	}
	pathutf16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	hr, _, _ := i.vtbl.CreateCookie.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(nameutf16)),
		uintptr(unsafe.Pointer(valueutf16)),
		uintptr(unsafe.Pointer(domainutf16)),
		uintptr(unsafe.Pointer(pathutf16)),
		uintptr(unsafe.Pointer(&cookie)),
	)
	if hr != 0 {
		return nil, syscall.Errno(hr)
	}
	return cookie, nil
}

// CopyCookie creates a copy of the given cookie
func (i *ICoreWebView2CookieManager) CopyCookie(cookie *ICoreWebView2Cookie) (*ICoreWebView2Cookie, error) {
	var newCookie *ICoreWebView2Cookie
	hr, _, _ := i.vtbl.CopyCookie.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(cookie)),
		uintptr(unsafe.Pointer(&newCookie)),
	)
	if hr != 0 {
		return nil, syscall.Errno(hr)
	}
	return newCookie, nil
}

// getCookiesWaiter collects the async GetCookies completion. The handler is
// held via waiter.handler, and the waiter via the package-level cookieWaiters
// map, so the Go object backing the raw COM pointer stays alive for the whole
// call (matching the fork's pattern of storing handlers on long-lived
// structs — see NewChromium).
type GetCookiesCompletion struct {
	key       uintptr
	done      chan struct{}
	errorCode uintptr
	list      *ICoreWebView2CookieList
	handler   *iCoreWebView2GetCookiesCompletedHandler
}

func (w *GetCookiesCompletion) QueryInterface(_, _ uintptr) uintptr { return 0 }
func (w *GetCookiesCompletion) AddRef() uintptr                      { return 1 }
func (w *GetCookiesCompletion) Release() uintptr                     { return 1 }

// GetCookiesCompleted is invoked by WebView2 on the UI thread (delivered by
// the app's running main message loop); the close(done) write publishes the
// results to the waiting goroutine.
func (w *GetCookiesCompletion) GetCookiesCompleted(errorCode uintptr, cookieList *ICoreWebView2CookieList) uintptr {
	w.errorCode = errorCode
	w.list = cookieList
	if cookieList != nil {
		cookieList.AddRef()
	}
	close(w.done)
	return 0
}

// cookieWaiterGraveyard parks waiters whose completion never arrived within
// the timeout. WebView2 may still invoke their handler later, so the backing
// Go memory must stay alive forever rather than be collected into a dangling
// COM pointer.
var (
	cookieWaiterGraveyardMu sync.Mutex
	cookieWaiterGraveyard   []*GetCookiesCompletion
)

// cookieWaiters tracks in-flight GetCookies calls keyed by the native manager
// pointer. A sync.Map — NOT a struct field — because ICoreWebView2CookieManager
// overlays a native COM object: writing Go fields into it corrupts native
// memory and crashes the process inside the next vtable call.
var cookieWaiters sync.Map // map[uintptr]*GetCookiesCompletion

// BeginGetCookies starts the asynchronous GetCookies call and returns a
// completion token.
//
// The ICoreWebView2CookieManager::GetCookies vtable call is apartment-affine:
// initiating it from a thread that is neither the WebView2 UI thread nor
// COM-initialized faults inside WebView2's proxy. Callers must invoke
// BeginGetCookies on the app main thread (e.g. inside InvokeSync); the
// completion handler is then delivered by the main message loop. Wait may be
// called from any goroutine.
func (i *ICoreWebView2CookieManager) BeginGetCookies(uri string) (*GetCookiesCompletion, error) {
	if i == nil {
		return nil, errors.New("GetCookies: nil manager")
	}
	if i.vtbl == nil {
		return nil, errors.New("GetCookies: nil vtable")
	}
	uriutf16, err := windows.UTF16PtrFromString(uri)
	if err != nil {
		return nil, err
	}

	key := uintptr(unsafe.Pointer(i))
	w := &GetCookiesCompletion{done: make(chan struct{}), key: key}
	w.handler = newICoreWebView2GetCookiesCompletedHandler(w)
	if _, loaded := cookieWaiters.LoadOrStore(key, w); loaded {
		return nil, errors.New("GetCookies: a call is already in flight on this manager")
	}

	hr, _, _ := i.vtbl.GetCookies.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(uriutf16)),
		uintptr(unsafe.Pointer(w.handler)),
	)
	runtime.KeepAlive(w.handler)
	if hr != 0 {
		cookieWaiters.Delete(key)
		return nil, syscall.Errno(hr)
	}
	return w, nil
}

// Wait blocks until the completion handler runs (on the UI thread's message
// loop) or the timeout elapses. On timeout the waiter is parked in the
// graveyard — its Go memory must outlive any late WebView2 callback — and the
// manager slot is freed so later calls are not blocked behind it.
func (w *GetCookiesCompletion) Wait(timeout time.Duration) (*ICoreWebView2CookieList, error) {
	defer cookieWaiters.Delete(w.key)
	select {
	case <-w.done:
	case <-time.After(timeout):
		cookieWaiterGraveyardMu.Lock()
		cookieWaiterGraveyard = append(cookieWaiterGraveyard, w)
		cookieWaiterGraveyardMu.Unlock()
		return nil, errors.New("GetCookies: timed out waiting for completion")
	}

	if w.errorCode != 0 {
		return nil, syscall.Errno(w.errorCode)
	}
	return w.list, nil
}

// GetCookies gets all cookies matching the URI, blocking until completion.
// The initiating COM call runs on the calling thread — which therefore must
// be the app main thread; use BeginGetCookies + Wait to marshal initiation
// from another goroutine.
func (i *ICoreWebView2CookieManager) GetCookies(uri string) (*ICoreWebView2CookieList, error) {
	w, err := i.BeginGetCookies(uri)
	if err != nil {
		return nil, err
	}
	return w.Wait(10 * time.Second)
}

// DeleteCookies deletes all cookies with matching name and uri
func (i *ICoreWebView2CookieManager) DeleteCookies(name, uri string) error {
	nameutf16, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	uriutf16, err := windows.UTF16PtrFromString(uri)
	if err != nil {
		return err
	}

	hr, _, _ := i.vtbl.DeleteCookies.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(nameutf16)),
		uintptr(unsafe.Pointer(uriutf16)),
	)
	if hr != 0 {
		return syscall.Errno(hr)
	}
	return nil
}

// DeleteCookiesWithDomainAndPath deletes all cookies matching the domain and path
func (i *ICoreWebView2CookieManager) DeleteCookiesWithDomainAndPath(domain, path string) error {
	domainutf16, err := windows.UTF16PtrFromString(domain)
	if err != nil {
		return err
	}
	pathutf16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	hr, _, _ := i.vtbl.DeleteCookiesWithDomainAndPath.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(domainutf16)),
		uintptr(unsafe.Pointer(pathutf16)),
	)
	if hr != 0 {
		return syscall.Errno(hr)
	}
	return nil
}

// AddOrUpdateCookie adds or updates a cookie
func (i *ICoreWebView2CookieManager) AddOrUpdateCookie(cookie *ICoreWebView2Cookie) error {
	hr, _, _ := i.vtbl.AddOrUpdateCookie.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(cookie)),
	)
	if hr != 0 {
		return syscall.Errno(hr)
	}
	return nil
}

// DeleteCookie deletes a specific cookie
func (i *ICoreWebView2CookieManager) DeleteCookie(cookie *ICoreWebView2Cookie) error {
	hr, _, _ := i.vtbl.DeleteCookie.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(cookie)),
	)
	if hr != 0 {
		return syscall.Errno(hr)
	}
	return nil
}

// DeleteAllCookies deletes all cookies
func (i *ICoreWebView2CookieManager) DeleteAllCookies() error {
	hr, _, _ := i.vtbl.DeleteAllCookies.Call(
		uintptr(unsafe.Pointer(i)),
	)
	if hr != 0 {
		return syscall.Errno(hr)
	}
	return nil
}