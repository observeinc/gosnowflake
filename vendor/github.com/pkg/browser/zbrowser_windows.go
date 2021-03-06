// Code generated by 'go generate'; DO NOT EDIT.

package browser

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modshell32 = windows.NewLazySystemDLL("shell32.dll")

	procShellExecuteW = modshell32.NewProc("ShellExecuteW")
)

func shellExecute(hwnd int, verb string, file string, args string, cwd string, showCmd int) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(verb)
	if err != nil {
		return
	}
	var _p1 *uint16
	_p1, err = syscall.UTF16PtrFromString(file)
	if err != nil {
		return
	}
	var _p2 *uint16
	_p2, err = syscall.UTF16PtrFromString(args)
	if err != nil {
		return
	}
	var _p3 *uint16
	_p3, err = syscall.UTF16PtrFromString(cwd)
	if err != nil {
		return
	}
	return _shellExecute(hwnd, _p0, _p1, _p2, _p3, showCmd)
}

func _shellExecute(hwnd int, verb *uint16, file *uint16, args *uint16, cwd *uint16, showCmd int) (err error) {
	r1, _, e1 := syscall.Syscall6(procShellExecuteW.Addr(), 6, uintptr(hwnd), uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(args)), uintptr(unsafe.Pointer(cwd)), uintptr(showCmd))
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}
