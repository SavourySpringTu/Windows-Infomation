package sysinfo

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	// IMAGE_FILE_MACHINE_* constants (from winnt.h)
	IMAGE_FILE_MACHINE_UNKNOWN = 0x0
	IMAGE_FILE_MACHINE_I386    = 0x014c // 32-bit
	IMAGE_FILE_MACHINE_AMD64   = 0x8664 // 64-bit
	IMAGE_FILE_MACHINE_ARM64   = 0xAA64 // ARM64
)

type SYSTEM_INFO struct {
	wProcessorArchitecture uint16
}

func GetOSInfo() (string, error) {
	var sysinfo SYSTEM_INFO

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetNativeSystemInfo := kernel32.NewProc("GetNativeSystemInfo")

	procGetNativeSystemInfo.Call(
		uintptr(unsafe.Pointer(&sysinfo)),
	)

	arch := ""
	switch sysinfo.wProcessorArchitecture {
	case 9:
		arch = "x64"
	case 0:
		arch = "x86"
	case 5:
		arch = "ARM"
	case 6:
		arch = "IA64"
	case 12:
		arch = "ARM64"
	default:
		arch = fmt.Sprintf("Unknown (%d)", sysinfo.wProcessorArchitecture)
	}

	return arch, nil
}

func GetOSInfo2() (string, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	isWow64Process2 := kernel32.NewProc("IsWow64Process2")
	getCurrentProcess := kernel32.NewProc("GetCurrentProcess")

	// Get handle to current process
	hProcess, _, _ := getCurrentProcess.Call()

	var processMachine uint16
	var nativeMachine uint16

	ret, _, err := isWow64Process2.Call(
		hProcess,
		uintptr(unsafe.Pointer(&processMachine)),
		uintptr(unsafe.Pointer(&nativeMachine)),
	)

	if ret == 0 {
		return "", err
	}

	// Map nativeMachine to OS architecture
	switch nativeMachine {
	case IMAGE_FILE_MACHINE_I386:
		return "32-bit (x86)", nil
	case IMAGE_FILE_MACHINE_AMD64:
		return "64-bit (x64)", nil
	case IMAGE_FILE_MACHINE_ARM64:
		return "64-bit (ARM64)", nil
	default:
		return fmt.Sprintf("Unknown architecture: 0x%x", nativeMachine), nil
	}
}
