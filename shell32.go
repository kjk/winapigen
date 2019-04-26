// THIS FILE WAS AUTO-GENERATED BY https://github.com/kjk/w/cmd/gengo
package w

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	libshell32 *windows.LazyDLL

	sHGetFolderPathW    *windows.LazyProc
	sHGetFolderLocation *windows.LazyProc
)

func init() {
	libshell32 = windows.NewLazySystemDLL("shell32.dll")
	sHGetFolderPathW = libshell32.NewProc("SHGetFolderPathW")
	sHGetFolderLocation = libshell32.NewProc("SHGetFolderLocation")
}

const (
	SHGFP_TYPE_CURRENT = 0
	SHGFP_TYPE_DEFAULT = 1
)

func SHGetFolderPathWSys(hwndOwner HWND, nFolder int32, hToken uintptr, dwFlags uint32, pszPath *WCHAR) HRESULT {
	ret, _, _ := syscall.Syscall6(sHGetFolderPathW.Addr(), 5,
		uintptr(hwndOwner),
		uintptr(nFolder),
		uintptr(hToken),
		uintptr(dwFlags),
		uintptr(unsafe.Pointer(pszPath)),
		0,
	)
	return HRESULT(ret)
}

func SHGetFolderLocationSys(hwndOwner HWND, nFolder uint32, hToken uintptr, dwReserved uint32, ppidl **ITEMIDLIST) HRESULT {
	ret, _, _ := syscall.Syscall6(sHGetFolderLocation.Addr(), 5,
		uintptr(hwndOwner),
		uintptr(nFolder),
		uintptr(hToken),
		uintptr(dwReserved),
		uintptr(unsafe.Pointer(ppidl)),
		0,
	)
	return HRESULT(ret)
}
