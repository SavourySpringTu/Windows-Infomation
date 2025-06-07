package appsInfo

import (
	"syscall"
	"time"
	"unsafe"
)

const (
	HKEY_LOCAL_MACHINE = 0x80000002
	READ_KEY           = 0x20019
)

var (
	advapi32             = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegEnumKeyExW    = advapi32.NewProc("RegEnumKeyExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procRegCloseKey      = advapi32.NewProc("RegCloseKey")
	path                 = `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`
)

type AppInfo struct {
	Name        string
	Version     string
	Publisher   string
	InstallDate string
}

func GetInfoApp() ([]AppInfo, error) {
	result := make([]AppInfo, 0)
	subKey, _ := syscall.UTF16PtrFromString(path)
	var hKey syscall.Handle

	ret, _, _ := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(READ_KEY),
		uintptr(unsafe.Pointer(&hKey)),
	)

	if ret != 0 {
		return result, nil
	}
	defer procRegCloseKey.Call(uintptr(hKey))

	var index uint32 = 0
	var data [256]uint16

	for {
		size := uint32(len(data))
		ret1, _, _ := procRegEnumKeyExW.Call(
			uintptr(hKey),
			uintptr(index),
			uintptr(unsafe.Pointer(&data[0])),
			uintptr(unsafe.Pointer(&size)),
			0, 0, 0, 0,
		)
		if ret1 != 0 {
			break
		}
		appKey := syscall.UTF16ToString(data[:size])
		infoApp, err := GetValue(path + `\` + appKey)
		if err == nil {
			result = append(result, infoApp)
		}
		index++
	}
	return result, nil
}
func GetValue(path string) (AppInfo, error) {
	var result AppInfo
	var hKey syscall.Handle
	keyPtr, _ := syscall.UTF16PtrFromString(path)
	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(keyPtr)),
		0,
		READ_KEY,
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procRegCloseKey.Call(uintptr(hKey))
	if ret != 0 {
		return result, err
	}
	name, errName := QueryGetValue("DisplayName", hKey)
	if errName == nil {
		version, _ := QueryGetValue("DisplayVersion", hKey)
		publisher, _ := QueryGetValue("Publisher", hKey)
		installDate, _ := QueryGetValue("InstallDate", hKey)
		inputLayout := "20060102"
		outputLayout := "02/01/2006"
		layout, _ := time.Parse(inputLayout, installDate)
		formattedDate := layout.Format(outputLayout)
		result = AppInfo{
			Name:        name,
			Version:     version,
			Publisher:   publisher,
			InstallDate: formattedDate,
		}
	} else {
		return result, errName
	}
	return result, nil
}

func QueryGetValue(name string, hKey syscall.Handle) (string, error) {
	var valName, _ = syscall.UTF16PtrFromString(name)
	var valType uint32
	buf := make([]uint16, 1)
	var size = uint32(len(buf))

	ret, _, err := procRegQueryValueExW.Call(
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
