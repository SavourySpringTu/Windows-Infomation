package sysinfo

import (
	"syscall"
	"unsafe"
)

var (
	advapi32             = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procCloseKey         = advapi32.NewProc("RegCloseKey")
)

const (
	HKEY_LOCAL_MACHINE = 0x80000002
	KEY_READ           = 0x20019
)

func GetInfoOSbyName(name string) (string, error) {
	path := `SOFTWARE\Microsoft\Windows NT\CurrentVersion`

	var hKey syscall.Handle
	subKey := syscall.StringToUTF16Ptr(path)

	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(KEY_READ),
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procCloseKey.Call(uintptr(unsafe.Pointer(&hKey)))
	if ret != 0 {
		return "", err
	}

	valName := syscall.StringToUTF16Ptr(name)
	var valType uint32
	var size uint32

	ret1, _, err1 := procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		0,
		uintptr(unsafe.Pointer(&size)),
	)

	buf := make([]uint16, size/2)
	ret1, _, err1 = procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret1 != 0 {
		return "", err1
	}
	return syscall.UTF16ToString(buf), nil
}
