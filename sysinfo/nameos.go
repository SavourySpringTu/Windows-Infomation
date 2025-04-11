package sysinfo

import (
	"fmt"
	"syscall"
	"unsafe"
)

func GetNameOS() {
	path := `SOFTWARE\Microsoft\Windows NT\CurrentVersion`

	var hKey syscall.Handle
	subKey := syscall.StringToUTF16Ptr(path)

	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExW = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procCloseKey = advapi32.NewProc("RegCloseKey")

	ret, _, _ := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(KEY_READ),
		uintptr(unsafe.Pointer(&hKey)),
	)

	if ret != 0 {
		fmt.Println("RegQueryValueExW failed:", syscall.Errno(ret))
	}

	valName := syscall.StringToUTF16Ptr("ProductName")
	var valType uint32
	var buf [256]uint16
	var size uint32 = uint32(len(buf))

	ret1, _, _ := procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret1 != 0 {
		fmt.Println("RegQueryValueExW failed:", syscall.Errno(ret1))
	}

	value := syscall.UTF16ToString(buf[:size])
	fmt.Println(value)
}
