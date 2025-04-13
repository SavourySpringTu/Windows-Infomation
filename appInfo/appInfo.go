package appInfo

import (
	"fmt"
	"syscall"
	"unsafe"
)

type AppInfo struct {
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
	path                 = `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`
)

const (
	HKEY_LOCAL_MACHINE = 0x80000002
	KEY_READ           = 0x20019
)

func GetAllAppInfo() ([]AppInfo, error) {
	var result []AppInfo
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
		app, _ := GetInfoApp(path + `\` + appKeyName)
		if app.Name != "" {
			result = append(result, app)
		}
		index++
	}
	return result, nil
}

func GetInfoApp(subKey string) (AppInfo, error) {
	appKey := syscall.StringToUTF16Ptr(subKey)
	var hKey syscall.Handle
	var result AppInfo
	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(appKey)),
		0,
		uintptr(KEY_READ),
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procCloseKey.Call(uintptr(hKey))
	if ret != 0 {
		return result, err
	}

	name, _ := queryReg(hKey, "DisplayName")
	version, _ := queryReg(hKey, "DisplayVersion")
	publisher, _ := queryReg(hKey, "Publisher")
	installdate, _ := queryReg(hKey, "InstallDate")

	result = AppInfo{
		Name:        name,
		Version:     version,
		Publisher:   publisher,
		InstallDate: installdate,
	}
	return result, nil
}

func queryReg(hKey syscall.Handle, name string) (string, error) {
	valName := syscall.StringToUTF16Ptr(name)
	var size uint32
	var valType uint32
	ret, _, err := procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 && ret != 234 {
		return "", err
	}

	buf := make([]uint16, size/2)
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
