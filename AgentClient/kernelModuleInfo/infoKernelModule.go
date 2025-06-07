package kernelModuleInfo

import (
	"errors"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	SC_MANAGER_ALL_ACCESS       = 0xF003F
	SERVICE_ALL_ACCESS          = 0xF01FF
	HKEY_LOCAL_MACHINE          = 0x80000002
	READ_KEY                    = 0x20019
	CALG_SHA_256                = 0x0000800c
	PROV_RSA_ARS                = 24
	CRYPT_VERIFYCONTEXT         = 0xF0000000
	HP_HASHVAL                  = 0x0002
	SystemModuleInformation     = 11
	STATUS_INFO_LENGTH_MISMATCH = 0xC0000004
)

const (
	path = `SYSTEM\CurrentControlSet\Services`
)

var (
	ntdll                        = syscall.NewLazyDLL("ntdll.dll")
	kernel32                     = syscall.NewLazyDLL("Kernel32.dll")
	advapi32                     = syscall.NewLazyDLL("Advapi32.dll")
	psapi                        = syscall.NewLazyDLL("Psapi.dll")
	procRegOpenKeyExW            = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW         = advapi32.NewProc("RegQueryValueExW")
	procRegEnumKeyExW            = advapi32.NewProc("RegEnumKeyExW")
	procOpenSCManagerW           = advapi32.NewProc("OpenSCManagerW")
	procOpenServiceW             = advapi32.NewProc("OpenServiceW")
	procQueryServiceStatusEx     = advapi32.NewProc("QueryServiceStatusEx")
	procRegCloseKey              = advapi32.NewProc("RegCloseKey")
	procCloseServiceHandle       = advapi32.NewProc("CloseServiceHandle")
	procCryptDestroyHash         = advapi32.NewProc("CryptDestroyHash")
	procCryptReleaseContext      = advapi32.NewProc("CryptReleaseContext")
	procCryptAcquireContextW     = advapi32.NewProc("CryptAcquireContextW")
	procCryptCreateHash          = advapi32.NewProc("CryptCreateHash")
	procCryptHashData            = advapi32.NewProc("CryptHashData")
	procCryptGetHashParam        = advapi32.NewProc("CryptGetHashParam")
	procGetSystemDirectoryW      = kernel32.NewProc("GetSystemDirectoryW")
	procEnumDeviceDrivers        = psapi.NewProc("EnumDeviceDrivers")
	procGetDeviceDriverBaseNameW = psapi.NewProc("GetDeviceDriverBaseNameW")
	procGetDeviceDriverFileNameW = psapi.NewProc("GetDeviceDriverFileNameW")
	procNtQuerySystemInformation = ntdll.NewProc("NtQuerySystemInformation")
)

type SERVICE_STATUS_PROCESS struct {
	dwServiceType             uint32
	dwCurrentState            uint32
	dwControlsAccepted        uint32
	dwWin32ExitCode           uint32
	dwServiceSpecificExitCode uint32
	dwCheckPoint              uint32
	dwWaitHint                uint32
	dwProcessId               uint32
	dwServiceFlags            uint32
}

var startupMode = []string{
	"BOOT_START",
	"SYSTEM_START",
	"AUTO_START",
	"DEMAND_START",
	"DISABLED",
}
var stateKernelModule = []string{
	"Access is denied",
	"SERVICE_STOPPED",
	"SERVICE_START_PENDING",
	"SERVICE_STOP_PENDING",
	"SERVICE_RUNNING",
	"SERVICE_CONTINUE_PENDING",
	"SERVICE_PAUSE_PENDING",
	"SERVICE_PAUSED",
}
var PathSystemROOT string

type RTL_PROCESS_MODULE_INFORMATION struct {
	Section          uintptr
	MappedBase       uintptr
	ImageBase        uintptr
	ImageSize        uint32
	Flags            uint32
	LoadOrderIndex   uint32
	InitOrderIndex   uint32
	LoadCount        uint32
	OffsetToFileName uint32
	FullPathName     [256]byte
}

type RTL_PROCESS_MODULES struct {
	NumberOfModules uint32
	Modules         [1]RTL_PROCESS_MODULE_INFORMATION
}

type KernelModuleInfo struct {
	Name        string
	Path        string
	SHA256      string
	StartupMode string
	State       string
}

func GetInfoKernelModule() ([]KernelModuleInfo, error) {
	var result []KernelModuleInfo
	var hKey syscall.Handle
	subKey, _ := syscall.UTF16PtrFromString(path)

	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(READ_KEY),
		uintptr(unsafe.Pointer(&hKey)),
	)
	defer procRegCloseKey.Call(uintptr(hKey))
	if ret != 0 {
		return result, err
	}
	ret, _, err = procOpenSCManagerW.Call(
		0,
		0,
		uintptr(SC_MANAGER_ALL_ACCESS),
	)
	defer procCloseServiceHandle.Call(ret)
	hManager := syscall.Handle(ret)
	if ret == 0 {
		return result, err
	}
	PathSystemROOT, _ = GetPathSystemRoot()

	// map kernel module
	mapKernel, _ := GetRunningKernelModules()

	for i := 0; ; i++ {
		var size = uint32(20)
		buf := make([]uint16, 20)

		ret, _, _ = procRegEnumKeyExW.Call(
			uintptr(hKey),
			uintptr(i),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0, 0, 0, 0,
		)
		name := syscall.UTF16ToString(buf[:])
		keyKernel := path + `\` + name
		typeKey, _ := GetValueKeyByName(keyKernel, "Type")
		if typeKey == "1" {
			//================== Get value ==========================
			imagePath, _ := GetValueKeyByName(keyKernel, "ImagePath")
			pathExe := ""
			if imagePath == "" {
				a := mapKernel[strings.ToLower(name)]
				pathExe = a
			} else {
				pathExe = strings.Split(imagePath, " ")[0]
			}
			absolutePathExe := ProcessStringSystemRoot(pathExe)

			indexModeStr, _ := GetValueKeyByName(keyKernel, "Start")
			indexMode, _ := strconv.Atoi(indexModeStr)
			stateIndex, _ := GetStateKernelModule(hManager, name)
			hash, _ := GetSHA256KernelModule(absolutePathExe)
			//================== Set value ==========================
			kernelModuleInfo := KernelModuleInfo{
				Name:        name,
				Path:        absolutePathExe,
				StartupMode: startupMode[indexMode],
				State:       stateKernelModule[stateIndex],
				SHA256:      hash,
			}
			result = append(result, kernelModuleInfo)
		}
		if ret == 259 {
			break
		}
	}
	return result, nil
}

func GetValueKeyByName(keyKernel string, name string) (string, error) {
	var hKey syscall.Handle
	subKey, _ := syscall.UTF16PtrFromString(keyKernel)
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
	buf := make([]uint16, 256)
	var size = uint32(len(buf))
	ret, _, _ = procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 234 && ret != 0 {
		return "", syscall.Errno(ret)
	}
	ret, _, _ = procRegQueryValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 {
		return "", syscall.Errno(ret)
	}
	if valType == syscall.REG_DWORD {
		val := *(*uint32)(unsafe.Pointer(&buf[0]))
		str := strconv.FormatUint(uint64(val), 10)
		return str, nil
	} else {
		return syscall.UTF16ToString(buf[:]), nil
	}
}

func GetStateKernelModule(hManager syscall.Handle, name string) (int, error) {
	namePtr, _ := syscall.UTF16PtrFromString(name)
	hService, _, errService := procOpenServiceW.Call(
		uintptr(hManager),
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(SERVICE_ALL_ACCESS),
	)
	defer procCloseServiceHandle.Call(hService)
	if hService == 0 {
		return 0, errService
	}
	var size = uint32(1000)
	var buf []byte
	var sizeRt uint32

	for {
		buf = make([]byte, size)
		ret, _, err := procQueryServiceStatusEx.Call(
			hService,
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret != 0 {
			break
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
			return 0, err
		}
		size = sizeRt
	}
	serviceStatus := (*SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0]))
	return int(serviceStatus.dwCurrentState), nil
}

func GetPathSystemRoot() (string, error) {
	size := uint32(20)
	buf := make([]uint16, size)
	ret, _, err := procGetSystemDirectoryW.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
	)
	size = uint32(ret + 1)
	buf = make([]uint16, size)
	ret, _, err = procGetSystemDirectoryW.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
	)
	if ret == 0 {
		return "", err
	}
	return syscall.UTF16ToString(buf[:]), nil
}

func ProcessStringSystemRoot(sysPath string) string {
	strPoint1 := "System32"
	strPoint2 := "system32"
	var strSplit string
	index := strings.Index(sysPath, strPoint1)
	if index != -1 {
		strSplit = sysPath[index+8:]
	} else {
		index = strings.Index(sysPath, strPoint2)
		if index != -1 {
			strSplit = sysPath[index+8:]
		} else {
			return sysPath
		}
	}
	result := PathSystemROOT + strSplit
	return result
}

// ====================== Get Running Kernel Moduels ====================

func GetRunningKernelModules() (map[string]string, error) {
	var result = make(map[string]string)

	var size = uint32(1)
	var buf = make([]uintptr, size)
	var sizeRt uint32
	ret, _, err := procEnumDeviceDrivers.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&sizeRt)),
	)
	if ret == 0 {
		return result, err
	}
	size = sizeRt
	buf = make([]uintptr, size)
	ret, _, err = procEnumDeviceDrivers.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&sizeRt)),
	)
	if ret == 0 || size == 0 || sizeRt == 0 {
		return result, err
	}
	count := int(sizeRt) / int((unsafe.Sizeof(buf[0])))
	for i := 0; i < count; i++ {
		name, errName := GetNameKernelModule(buf[i])
		if errName != nil {
			continue
		}
		pathName, _ := GetPathKernelModule(buf[i])
		if pathName != "" {
			result[strings.ToLower(name)] = pathName
		}

	}
	return result, nil
}

// Get Name kernel modulle
func GetNameKernelModule(imageBase uintptr) (string, error) {
	var size = uint32(10)
	var name = make([]uint16, size)
	for {
		ret, _, _ := procGetDeviceDriverBaseNameW.Call(
			uintptr(imageBase),
			uintptr(unsafe.Pointer(&name[0])),
			uintptr(size),
		)
		if size > uint32(ret) {
			break
		}
		if size == 0 || ret == 0 {
			size = 0
			break
		}
		size = size * 2
		name = make([]uint16, size)
	}
	if size == 0 {
		return "", errors.New("GetDeciveDriverBaseName fail!")
	}
	result := cutUTF16BeforeDot(name[:])
	return syscall.UTF16ToString(result[:]), nil
}

func GetPathKernelModule(imageBase uintptr) (string, error) {
	var size = uint32(10)
	var pathName = make([]uint16, size)
	for {
		ret, _, _ := procGetDeviceDriverFileNameW.Call(
			uintptr(imageBase),
			uintptr(unsafe.Pointer(&pathName[0])),
			uintptr(size),
		)
		if size > uint32(ret) {
			break
		}
		if size == 0 || ret == 0 {
			size = 0
			break
		}
		size = size * 2
		pathName = make([]uint16, size)
	}
	if size == 0 {
		return "", errors.New("GetDeviceDriverFileName fail!")
	}
	return syscall.UTF16ToString(pathName[:]), nil
}

func cutUTF16BeforeDot(buf []uint16) []uint16 {
	for i, r := range buf {
		if r == '.' {
			return buf[:i]
		}
	}
	return buf
}
