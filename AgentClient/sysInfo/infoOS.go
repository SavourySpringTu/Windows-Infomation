package sysInfo

import (
	"strconv"
	"syscall"
	"unsafe"
)

const (
	HKEY_LOCAL_MACHINE = 0x80000002
	READ_KEY           = 0x20019
)

var (
	ntdll                = syscall.NewLazyDLL("Ntdll.dll")
	procRtlGetVersion    = ntdll.NewProc("RtlGetVersion")
	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procRegCloseKey      = advapi32.NewProc("RegCloseKey")
	pathInfoOS           = `SOFTWARE\Microsoft\Windows NT\CurrentVersion`
)

type osversioninfoexw struct {
	dwOSVersionInfoSize uint32
	dwMajorVersion      uint32
	dwMinorVersion      uint32
	dwBuildNumber       uint32
	dwPlatformId        uint32
	szCSDVersion        [128]uint16
	wServicePackMajor   uint16
	wServicePackMinor   uint16
	wSuiteMask          uint16
	wProductType        uint8
	wReserved           uint8
}

func GetNameOS() (string, error) {
	var osversion osversioninfoexw

	_, _, err := procRtlGetVersion.Call(
		uintptr(unsafe.Pointer(&osversion)),
	)

	if err.Error() != "The operation completed successfully." {
		return "", err
	}

	var win string
	if int(osversion.dwMajorVersion) == 5 {
		if int(osversion.dwMinorVersion) == 0 {
			win = "2000"
		} else if int(osversion.dwMinorVersion) == 1 {
			win = "XP"
		} else {
			win = "Server 2003"
		}
	} else if int(osversion.dwMajorVersion) == 6 {
		if int(osversion.dwMinorVersion) == 1 {
			win = "7"
		} else if int(osversion.dwMinorVersion) == 2 {
			win = "8"
		} else if int(osversion.dwMinorVersion) == 3 {
			win = "8.1"
		}
	} else {
		if int(osversion.dwBuildNumber) < 22000 {
			win = "10"
		} else {
			win = "11"
		}
	}
	result := "Windows " + win + " Build " + strconv.Itoa(int(osversion.dwBuildNumber))
	return result, nil
}

// ========================== LAY TRONG REGISTRY ================================================

func GetInfoOSbyName(name string) (string, error) {
	subKey, _ := syscall.UTF16PtrFromString(pathInfoOS)
	var hKey syscall.Handle

	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(READ_KEY),
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procRegCloseKey.Call(uintptr(hKey))
	if ret != 0 {
		return "", err
	}

	var valName, _ = syscall.UTF16PtrFromString(name)
	var valType uint32
	buf := make([]uint16, 1)
	var size = uint32(len(buf))
	ret, _, err = procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return syscall.UTF16ToString(buf), nil
	} else if ret == 234 {
		buf = make([]uint16, size/2)
		ret, _, err = procRegQueryValueExW.Call(
			uintptr(hKey),
			uintptr(unsafe.Pointer(valName)),
			0,
			uintptr(unsafe.Pointer(&valType)),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret != 0 {
			return "", err
		}
		return syscall.UTF16ToString(buf), nil
	}
	return "", err
}
