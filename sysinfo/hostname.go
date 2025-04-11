package sysinfo

import (
	"syscall"
	"unsafe"
)

func GetUsername() (string, error) {
	var buf [256]uint16
	size := uint32(len(buf))

	modadvapi := syscall.NewLazyDLL("advapi32.dll")
	procGetUserName := modadvapi.NewProc("GetUserNameW")

	ret, _, err := procGetUserName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == 0 {
		return "", err
	}

	return syscall.UTF16ToString(buf[:size]), nil
}
