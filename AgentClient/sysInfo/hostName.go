package sysInfo

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	advapi32         = syscall.NewLazyDLL("Advapi32.dll")
	procGetUserNameW = advapi32.NewProc("GetUserNameW")
)

func GetUserName() (string, error) {
	var size uint32
	ret, _, err := procGetUserNameW.Call(
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 && errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == true {
		buf := make([]uint16, size)

		ret, _, err = procGetUserNameW.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret == 0 {
			return "", err
		}
		return syscall.UTF16ToString(buf), nil
	}
	return "", err
}
