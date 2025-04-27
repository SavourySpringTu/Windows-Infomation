package kernelModuleInfo

import (
	"fmt"
	"main/global"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	advapi               = syscall.NewLazyDLL("Advapi32.dll")
	procRegOpenKeyExW    = advapi.NewProc("RegOpenKeyExW")
	procRegEnumKeyExW    = advapi.NewProc("RegEnumKeyExW")
	procRegQueryValueExW = advapi.NewProc("RegQueryValueExW")
	procRegCloseKey      = advapi.NewProc("RegCloseKey")
)

const (
	pathKey = `SYSTEM\CurrentControlSet\Services`
)

var Start = []string{"Boot", "System", "Automatic", "Manual", "Disable"}

type KernelModuleInfo struct {
	Name    string
	Version string
	Path    string
	Status  string
}

func GetKernelModuleInfo() ([]KernelModuleInfo, error) {
	var result []KernelModuleInfo
	var hKey syscall.Handle
	pathKeyPtr, _ := syscall.UTF16PtrFromString(pathKey)
	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(global.HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(pathKeyPtr)),
		0,
		uintptr(global.KEY_READ),
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procRegCloseKey.Call(uintptr(hKey))

	if ret != 0 {
		return result, err
	}

	var index uint32 = 0
	var nameBuf [256]uint16
	for {
		nameLen := uint32(len(nameBuf))
		ret, _, err = procRegEnumKeyExW.Call(
			uintptr(hKey),
			uintptr(index),
			uintptr(unsafe.Pointer(&nameBuf[0])),
			uintptr(unsafe.Pointer(&nameLen)),
			0, 0, 0, 0,
		)
		if ret != 0 {
			break
		}

		serviceName := syscall.UTF16ToString(nameBuf[:nameLen])
		moduleInfo, err := getModuleInfo(hKey, serviceName)
		if err == nil {
			result = append(result, moduleInfo)
		}
		index++
	}
	return result, nil
}

func getModuleInfo(hKey syscall.Handle, serviceName string) (KernelModuleInfo, error) {
	var result KernelModuleInfo
	result.Name = serviceName

	var serviceKey syscall.Handle
	serviceNamePtr, _ := syscall.UTF16PtrFromString(serviceName)
	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(serviceNamePtr)),
		0,
		uintptr(global.KEY_READ),
		uintptr(unsafe.Pointer(&serviceKey)),
	)
	defer procRegCloseKey.Call(uintptr(serviceKey))

	if ret != 0 {
		return result, err
	}

	typeStr, err := queryRegValue(serviceKey, "Type")
	if err != nil {
		return result, err
	}

	typeInt, err := strconv.Atoi(typeStr)
	if err != nil || typeInt != 1 {
		return result, err
	}
	version, _ := queryRegValue(serviceKey, "Version")
	result.Version = version

	path, _ := queryRegValue(serviceKey, "ImagePath")
	result.Path = path

	status, _ := queryRegValue(serviceKey, "Start")
	result.Status = getStatusFromStartValue(status)

	return result, nil
}

func queryRegValue(hKey syscall.Handle, valueName string) (string, error) {
	valName, _ := syscall.UTF16PtrFromString(valueName)
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
	if ret != 0 && ret != global.ERROR_MORE_DATA {
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
	fmt.Println(syscall.UTF16ToString(buf))
	return syscall.UTF16ToString(buf), nil
}

func getStatusFromStartValue(startValue string) string {
	switch startValue {
	case "0":
		return "Boot"
	case "1":
		return "System"
	case "2":
		return "Automatic"
	case "3":
		return "Manual"
	case "4":
		return "Disabled"
	default:
		return "Unknown"
	}
}
