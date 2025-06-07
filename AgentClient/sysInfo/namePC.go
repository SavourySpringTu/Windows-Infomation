package sysInfo

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("Kernel32.dll")
	procGetComputerNameW = kernel32.NewProc("GetComputerNameW")
)

type SystemInfo struct {
	NamePC   string
	NameOs   string
	HostName string
	TimeUp   int
}

func GetSystemInfo() (SystemInfo, error) {
	namepc, _ := GetNamePC()
	nameos, _ := GetNameOS()
	hostname, _ := GetUserName()
	timeup, _ := GetTimeUp()
	sysinfo := SystemInfo{
		NamePC:   namepc,
		NameOs:   nameos,
		HostName: hostname,
		TimeUp:   timeup,
	}
	return sysinfo, nil
}

func GetNamePC() (string, error) {
	var size uint32
	ret, _, err := procGetComputerNameW.Call(
		0,
		uintptr(unsafe.Pointer(&size)),
	)

	if err != nil && errors.Is(err, syscall.ERROR_BUFFER_OVERFLOW) == true {
		buf := make([]uint16, size)
		ret, _, err = procGetComputerNameW.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret == 0 {
			return "", err
		}
		return syscall.UTF16ToString(buf), nil
	} else {
		return "", err
	}
}
