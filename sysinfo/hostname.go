package sysinfo

import (
	"syscall"
	"unsafe"
)

var (
	modadvapi       = syscall.NewLazyDLL("advapi32.dll")
	procGetUserName = modadvapi.NewProc("GetUserNameW")
)

func GetUsername() (string, error) {
	var size uint32

	ret, _, err := procGetUserName.Call(
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	buf := make([]uint16, size/2)
	if ret == 234 {
		ret, _, err := procGetUserName.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret == 0 {
			return "", err
		}
	} else {
		return "", err
	}
	return syscall.UTF16ToString(buf), nil
}
