package sysinfo

import (
	"syscall"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount64 = kernel32.NewProc("GetTickCount64")
)

func GetUptime() (int, error) {
	ret, _, _ := procGetTickCount64.Call()

	milliseconds := int(ret) / 1000 / 60

	return milliseconds, nil
}
