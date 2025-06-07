package sysInfo

import (
	"unsafe"
)

type system_info struct {
	wProcessorArchitecture      uint16
	wReserved                   uint16
	dwPageSize                  uint32
	lpMinimumApplicationAddress unsafe.Pointer
	lpMaximumApplicationAddress unsafe.Pointer
	dwActiveProcessorMask       uintptr
	dwNumberOfProcessors        uint32
	dwProcessorType             uint32
	dwAllocationGranularity     uint32
	wProcessorLevel             uint16
	wProcessorRevision          uint16
}

var architectureMap = map[int]string{
	0:  "(x86) 32 bit",
	5:  "(ARM) 64 bit",
	6:  "(IA64) IA 64",
	9:  "(x64) 64 bit",
	12: "(ARM64) 64 bit",
}
var (
	procGetSystemInfo = kernel32.NewProc("GetNativeSystemInfo")
)

func GetArchitectureOS() (string, error) {
	var sysinfo system_info

	_, _, err := procGetSystemInfo.Call(
		uintptr(unsafe.Pointer(&sysinfo)),
	)
	if err.Error() != "The operation completed successfully." {
		return "", err
	}

	if sysinfo.wProcessorArchitecture == 0xFFFF {
		return "Unknown", nil
	} else {
		return architectureMap[int(sysinfo.wProcessorArchitecture)], nil
	}
}
