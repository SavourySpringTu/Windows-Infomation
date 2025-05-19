package sysInfo

import (
	"errors"
	"fmt"
	"main/global"
	"syscall"
	"unsafe"
)

var (
	advapi32                   = syscall.NewLazyDLL("advapi32.dll")
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procGetUserName            = advapi32.NewProc("GetUserNameW")
	procRegOpenKeyExW          = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW       = advapi32.NewProc("RegQueryValueExW")
	procGetTickCount64         = kernel32.NewProc("GetTickCount64")
	procGetSystemFirmwareTable = kernel32.NewProc("GetSystemFirmwareTable")
	procCloseKey               = advapi32.NewProc("RegCloseKey")
)

type RawSMBIOSData struct {
	Used20CallingMethod byte
	SMBIOSMajorVersion  byte
	SMBIOSMinorVersion  byte
	DmiRevision         byte
	Length              uint32
	SMBIOSTableData     []byte
}

type HeaderSMBIOS struct {
	Type   byte
	Length byte
	Handle uint16
}

func GetInfoOSbyName(name string) (string, error) {
	path := `SOFTWARE\Microsoft\Windows NT\CurrentVersion`
	var hKey syscall.Handle
	subKey, _ := syscall.UTF16PtrFromString(path)

	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(global.HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(global.KEY_READ),
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procCloseKey.Call(uintptr(unsafe.Pointer(&hKey)))
	if ret != 0 {
		return "", err
	}

	valName, _ := syscall.UTF16PtrFromString(name)
	var valType uint32
	var size uint32
	buf := make([]uint16, 1)
	ret, _, err = procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == global.ERROR_MORE_DATA {
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
	} else if ret != 0 && ret != global.ERROR_MORE_DATA {
		return "", err
	}
	return syscall.UTF16ToString(buf), nil
}
func GetUsername() (string, error) {
	var size uint32

	buf := make([]uint16, 3)
	ret, _, err := procGetUserName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == true {
		buf = make([]uint16, size)
		ret, _, err = procGetUserName.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret == 0 {
			return "", err
		}
	} else if ret == 0 {
		return "", err
	}
	return syscall.UTF16ToString(buf), nil
}
func GetUptime() (int, error) {
	ret, _, _ := procGetTickCount64.Call()
	milliseconds := int(ret) / 1000 / 60
	return milliseconds, nil
}

func GetFirmwareSystem() {
	var size = uint32(32)
	var buf = make([]byte, size)
	ret, _, _ := procGetSystemFirmwareTable.Call(
		global.RSMB,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
	)
	size = uint32(ret)
	buf = make([]byte, size)
	ret, _, _ = procGetSystemFirmwareTable.Call(
		global.RSMB,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
	)
	smDataTable := buf[8:size]

	table, _ := GetTableByType(smDataTable, 1)
	header := (*HeaderSMBIOS)(unsafe.Pointer(&table[0]))
	fmt.Println(header.Length)
	fmt.Println(header.Type)
}

func GetTableByType(buf []byte, typeTable int) ([]byte, error) {
	start := 0
	for i := 0; i < len(buf)-1; i++ {
		fmt.Println(i)
		if buf[i] == 0x00 && buf[i+1] == 0x00 {
			fmt.Println("i: ", i)
			table := buf[start:]
			header := (*HeaderSMBIOS)(unsafe.Pointer(&table[0]))
			if int(header.Type) == typeTable {
				return table, nil
			}
			start = i + 2
		}
	}
	return nil, errors.New("Can't find")
}

func getSMBiosStringByFormattedIndex(buffString []byte, index byte) string {
	if index == 0 {
		return ""
	}

	currentIndex := byte(1)
	start := 0

	for i := 0; i < len(buffString); i++ {
		if buffString[i] == 0 {
			if currentIndex == index {
				return string(buffString[start:i])
			}
			currentIndex++
			start = i + 1

			if i+1 < len(buffString) && buffString[i+1] == 0 {
				break
			}
		}
	}
	return ""
}
