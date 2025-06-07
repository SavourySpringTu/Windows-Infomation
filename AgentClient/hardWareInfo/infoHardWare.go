package hardWareInfo

import (
	"errors"
	"syscall"
	"unsafe"
)

type MEMORYSTATUSEX struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

const (
	MAX_PATH                      = 260
	GENERIC_ALL                   = 0x10000000
	FILE_SHARE_READ               = 0x00000001
	HKEY_LOCAL_MACHINE            = 0x80000002
	READ_KEY                      = 0x20019
	GENERIC_READ                  = 0x80000000
	GENERIC_WRITE                 = 0x40000000
	OPEN_EXISTING                 = 3
	FILE_ATTRIBUTE_NORMAL         = 0x80
	IOCTL_DISK_GET_DRIVE_GEOMETRY = 0x0007000000000028
)

type DISK_GEOMETRY struct {
	Cylinders         uint64
	MediaType         uint32
	TracksPerCylinder uint32
	SectorsPerTrack   uint32
	BytesPerSector    uint32
}

var (
	kernel32                 = syscall.NewLazyDLL("Kernel32.dll")
	advapi32                 = syscall.NewLazyDLL("Advapi32.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpaceExW  = kernel32.NewProc("GetDiskFreeSpaceExW")
	procRegOpenKeyExW        = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW     = advapi32.NewProc("RegQueryValueExW")
	procRegCloseKey          = advapi32.NewProc("RegCloseKey")
	procFindFirstVolumeW     = kernel32.NewProc("FindFirstVolumeW")
	procFindNextVolumeW      = kernel32.NewProc("FindNextVolumeW")
	procCreateFileW          = kernel32.NewProc("CreateFileW")
	procDeviceIoControl      = kernel32.NewProc("DeviceIoControl")
	procCloseHandle          = kernel32.NewProc("CloseHandle")
	pathCentralProcessor     = `HARDWARE\DESCRIPTION\System\CentralProcessor\0`
)

type HardWareInfo struct {
	NameCPU  string
	RAM      uint64
	SizeDisk uint64
}

func GetHardWareInfo() (HardWareInfo, error) {
	cpu, _ := GetNameCPU()
	ram, _ := GetInfoRAM()
	disk, _ := GetSizeDisk()
	result := HardWareInfo{
		NameCPU:  cpu,
		RAM:      ram,
		SizeDisk: disk,
	}
	return result, nil
}

func GetNameCPU() (string, error) {
	var hKey syscall.Handle
	subKey, _ := syscall.UTF16PtrFromString(pathCentralProcessor)
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

	var valName, _ = syscall.UTF16PtrFromString("ProcessorNameString")
	var valType uint32
	var buf = make([]uint16, 1)
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
		ret, _, _ = procRegQueryValueExW.Call(
			uintptr(hKey),
			uintptr(unsafe.Pointer(valName)),
			0,
			uintptr(unsafe.Pointer(&valType)),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret != 0 {
			err = errors.New("Error RegQueryValueExW")
			return "", err
		}
		return syscall.UTF16ToString(buf), nil
	}
	err = errors.New("Error RegQueryValueExW")
	return "", err
}

func GetInfoRAM() (uint64, error) {
	var memory MEMORYSTATUSEX
	memory.dwLength = uint32(unsafe.Sizeof(memory))
	ret, _, err := procGlobalMemoryStatusEx.Call(
		uintptr(unsafe.Pointer(&memory)),
	)
	if ret == 0 {
		return 0, err
	}
	return memory.ullTotalPhys / 1000000, nil
}

func GetSizeDisk() (uint64, error) {
	buf := make([]uint16, 49)
	size := uint32(len(buf))

	ret, _, err := procFindFirstVolumeW.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return 0, err
	}
	hFile := syscall.Handle(ret)
	defer procCloseHandle.Call(uintptr(hFile))

	result := uint64(0)
	for {
		sizeVolume, _ := GetSizeVolume(syscall.UTF16ToString(buf[:]))
		result = result + sizeVolume
		retNext, _, errNext := procFindNextVolumeW.Call(
			uintptr(hFile),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)

		if retNext == 0 {
			if errors.Is(errNext, syscall.ERROR_NO_MORE_FILES) == true {
				break
			} else {
				return result, errNext
			}
		}
	}
	return result, nil
}
func GetSizeVolume(pathDisk string) (uint64, error) {
	var freeBytesAvailableToCaller uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64
	ptrPathDisk, _ := syscall.UTF16PtrFromString(pathDisk)
	ret, _, err := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(ptrPathDisk)),
		uintptr(unsafe.Pointer(&freeBytesAvailableToCaller)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if ret == 0 {
		return 0, err
	}
	return totalNumberOfBytes / 1000000, nil
}
