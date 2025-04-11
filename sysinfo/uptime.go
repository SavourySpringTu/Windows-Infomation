package sysinfo

import (
	"syscall"
	"time"
)

func GetUptime() (time.Duration, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount64 := kernel32.NewProc("GetTickCount64")

	ret, _, _ := procGetTickCount64.Call()

	milliseconds := uint64(ret)

	return time.Duration(milliseconds) * time.Millisecond, nil
}
