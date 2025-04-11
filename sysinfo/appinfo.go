package sysinfo

import (
	"fmt"
	"syscall"
	"unsafe"
)

type info struct {
	Name        string
	Version     string
	Publisher   string
	InstallDate string
}

var (
	advapi32             = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procCloseKey         = advapi32.NewProc("RegCloseKey")
	procRegEnumKeyExW    = advapi32.NewProc("RegEnumKeyExW")
)

const (
	HKEY_LOCAL_MACHINE = 0x80000002
	KEY_READ           = 0x20019
)

func GetAppInfo() {
	path := `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`
	subKey := syscall.StringToUTF16Ptr(path)
	var hKey syscall.Handle

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

	var index uint32 = 0
	var nameBuf [256]uint16
	for {
		nameLen := uint32(len(nameBuf))
		ret, _, _ := procRegEnumKeyExW.Call(
			uintptr(hKey),
			uintptr(index),
			uintptr(unsafe.Pointer(&nameBuf[0])),
			uintptr(unsafe.Pointer(&nameLen)),
			0, 0, 0, 0,
		)
		if ret != 0 {
			break
		}
		appKeyName := syscall.UTF16ToString(nameBuf[:nameLen])
		abc(path + `\` + appKeyName)

		index++
	}
}

func abc(subKey string) {
	appKey := syscall.StringToUTF16Ptr(subKey)
	var hKey syscall.Handle
	ret, _, _ := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(appKey)),
		0,
		uintptr(KEY_READ),
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procCloseKey.Call(uintptr(hKey))
	if ret != 0 {
		fmt.Println("RegQueryValueExW failed:", syscall.Errno(ret))
	}

	name := queryReg(hKey, "DisplayName")
	version := queryReg(hKey, "DisplayVersion")
	publisher := queryReg(hKey, "Publisher")
	installdate := queryReg(hKey, "InstallDate")

	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Publisher: %s\n", publisher)
	fmt.Printf("Install Date: %s\n\n", installdate)
}

func queryReg(hKey syscall.Handle, name string) string {
	valName := syscall.StringToUTF16Ptr(name)
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
		return ""
	}

	value := syscall.UTF16ToString(buf[:size])
	return value
}
