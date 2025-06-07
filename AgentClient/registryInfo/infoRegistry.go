package registryInfo

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"
)

const (
	ERROR_NO_MORE_ITEMS = 259
	ERROR_MORE_DATA     = 234
	READ_KEY            = 0x20019
	HKEY_LOCAL_MACHINE  = 0x80000002
)

var (
	advapi32          = syscall.NewLazyDLL("Advapi32.dll")
	procRegOpenKeyExW = advapi32.NewProc("RegOpenKeyExW")
	procRegEnumValueW = advapi32.NewProc("RegEnumValueW")
	procRegCloseKey   = advapi32.NewProc("RegCloseKey")
)

type RegistryInfo struct {
	Name string
	Data string
}

func GetInfoRegistryByPath(path string) ([]RegistryInfo, error) {
	var result []RegistryInfo
	pathConvert, _, errConvert := CheckPathAndConvert(path)
	if errConvert != nil {
		return result, errConvert
	}

	subKey, _ := syscall.UTF16PtrFromString(pathConvert)
	var hKey syscall.Handle
	ret, _, _ := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(READ_KEY),
		uintptr(unsafe.Pointer(&hKey)),
	)

	if ret != 0 {
		return result, errors.New("Open key fail!")
	}
	defer procRegCloseKey.Call(uintptr(hKey))

	for index := 0; ; index++ {
		var name = make([]uint16, 30)
		var data = make([]uint16, 30)
		var nameSize = uint32(len(name))
		var dataSize = uint32(len(data))
		ret, _, _ = procRegEnumValueW.Call(
			uintptr(hKey),
			uintptr(index),
			uintptr(unsafe.Pointer(&name[0])),
			uintptr(unsafe.Pointer(&nameSize)),
			0, 0,
			uintptr(unsafe.Pointer(&data[0])),
			uintptr(unsafe.Pointer(&dataSize)),
		)

		if ret == ERROR_NO_MORE_ITEMS {
			break
		} else if ret != 0 && ret != ERROR_MORE_DATA {
			continue
		} else if ret == ERROR_MORE_DATA {
			name = make([]uint16, nameSize+1)
			data = make([]uint16, dataSize+1)
			nameSize++
			dataSize++
			ret, _, _ = procRegEnumValueW.Call(
				uintptr(hKey),
				uintptr(index),
				uintptr(unsafe.Pointer(&name[0])),
				uintptr(unsafe.Pointer(&nameSize)),
				0, 0,
				uintptr(unsafe.Pointer(&data[0])),
				uintptr(unsafe.Pointer(&dataSize)),
			)
			if ret != 0 {
				continue
			}
		}
		if syscall.UTF16ToString(name) != "" {
			regisInfo := RegistryInfo{
				Name: syscall.UTF16ToString(name),
				Data: syscall.UTF16ToString(data),
			}
			result = append(result, regisInfo)
		}
	}
	return result, nil
}

func CheckPathAndConvert(path string) (string, int, error) {
	hkey := 0
	if strings.HasPrefix(path, "Computer") == true {
		path = path[9:]
	}
	if strings.HasPrefix(path, "HKEY_LOCAL_MACHINE") == true {
		//hkey = HKEY_LOCAL_MACHINE
		path = path[19:]
	}
	//else if strings.HasPrefix(path, "HKEY_CURRENT_CONFIG") == true {
	//	hKey = syscall.HKEY_CURRENT_CONFIG
	//	path = path[20:]
	//} else if strings.HasPrefix(path, "HKEY_USERS") == true {
	//	hKey = syscall.HKEY_USERS
	//	path = path[11:]
	//} else if strings.HasPrefix(path, "HKEY_CURRENT_USER") == true {
	//	hKey = syscall.HKEY_CURRENT_USER
	//	path = path[18:]
	//} else if strings.HasPrefix(path, "HKEY_CLASSES_ROOT") == true {
	//	hKey = syscall.HKEY_CLASSES_ROOT // 0x80000000
	//	path = path[18:]
	//}
	return path, hkey, nil
}
