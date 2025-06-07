package processesInfo

import (
	"AgentClient/connectionsInfo"
	"bytes"
	"errors"
	"golang.org/x/sys/windows"
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("Kernel32.dll")
	ntdll                          = syscall.NewLazyDLL("ntdll.dll")
	advapi32                       = syscall.NewLazyDLL("advapi32.dll")
	psapi                          = syscall.NewLazyDLL("Psapi.dll")
	procOpenProcess                = kernel32.NewProc("OpenProcess")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetProcessTimes            = kernel32.NewProc("GetProcessTimes")
	procFileTimeToSystemTime       = kernel32.NewProc("FileTimeToSystemTime")
	procLookupAccountSidW          = advapi32.NewProc("LookupAccountSidW")
	procCreateToolhelp32Snapshot   = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First             = kernel32.NewProc("Process32First")
	procProcess32Next              = kernel32.NewProc("Process32Next")
	procModule32FirstW             = kernel32.NewProc("Module32FirstW")
	procModule32NextW              = kernel32.NewProc("Module32NextW")
	procReadProcessMemory          = kernel32.NewProc("ReadProcessMemory")
	procNtQueryInformationProcess  = ntdll.NewProc("NtQueryInformationProcess")
	procZwQueryInformationProcess  = ntdll.NewProc("ZwQueryInformationProcess")
	procCloseHandle                = kernel32.NewProc("CloseHandle")
)

const (
	ProcessCommandLineInformation = 60
	ProcessProtectionInformation  = 61
	ProcessBasicInformation       = 0
)
const (
	TH32CS_SNAPPROCESS                = 0x00000002
	TH32CS_SNAPMODULE                 = 0x00000008
	TH32CS_SNAPMODULE32               = 0x00000010
	PROCESS_ALL_ACCESS                = 0x000F0000
	PROCESS_QUERY_INFORMATION         = 0x0400
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000

	PROCESS_VM_READ             = 0x0010
	STATUS_INFO_LENGTH_MISMATCH = 0xC0000004
	STATUS_SUCCESS              = 0x00000000
	LIST_MODULES_ALL            = 0x03
	ERROR_PARTIAL_COPY          = 0x12B
)
const (
	MAX_PATH          = 260
	MAX_MODULE_NAME32 = 255
)

type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [MAX_PATH]byte
}

type MODULEENTRY32 struct {
	Size         uint32
	ModuleID     uint32
	ProcessID    uint32
	GlobalUsage  uint32
	ProccntUsage uint32
	BaseAddr     uintptr
	BaseSize     uint32
	hModule      uintptr
	szModule     [MAX_MODULE_NAME32 + 1]uint16
	szExePath    [MAX_PATH]uint16
}

type FILETIME struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

type SYSTEMTIME struct {
	wYear         uint16
	wMonth        uint16
	wDayOfWeek    uint16
	wDay          uint16
	wHour         uint16
	wMinute       uint16
	wSecond       uint16
	wMilliseconds uint16
}

type PS_PROTECTION struct {
	Level uint8
}

func (p *PS_PROTECTION) Type() uint8 {
	return p.Level & 0x07
}

func (p *PS_PROTECTION) Audit() uint8 {
	return p.Level >> 3 & 0x01
}

func (p *PS_PROTECTION) Signer() uint8 {
	return p.Level >> 4 & 0x0F
}

type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

//================================================================

type ProcessInfo struct {
	Pid           uint32
	ParentPid     uint32
	Runtime       uint32
	Name          string
	Path          string
	Commandline   string
	Token         TokenProcess
	ConnectionTcp []connectionsInfo.ConnectionTcpInfo
	ConnectionUdp []connectionsInfo.ConnectionUdpInfo
	Module        []moduleInfo
}

type moduleInfo struct {
	Name string
}

var (
	ProtectedSignerNames = []string{
		"No",
		"PsProtectedSignerAuthenticode",
		"PsProtectedSignerCodeGen",
		"PsProtectedSignerAntimalware",
		"PsProtectedSignerLsa",
		"PsProtectedSignerWindows",
		"PsProtectedSignerWinTcb",
		"PsProtectedSignerWinSystem",
		"PsProtectedSignerApp",
		"PsProtectedSignerMax",
	}
)

func GetInfoProcesses(pid uint32, all bool) ([]ProcessInfo, error) {
	var result []ProcessInfo
	var processEntry PROCESSENTRY32
	hProcessSnap, _, err := procCreateToolhelp32Snapshot.Call(
		TH32CS_SNAPPROCESS,
		0,
	)
	if hProcessSnap == 0 {
		return result, err
	}

	processEntry.DwSize = uint32(unsafe.Sizeof(processEntry))
	ret, _, err := procProcess32First.Call(
		hProcessSnap,
		uintptr(unsafe.Pointer(&processEntry)),
	)
	if ret == 0 {
		return result, err
	}
	// Get map connections
	mapConnectionTcp, _ := connectionsInfo.GetTcpInfo()
	mapConnectionUdp, _ := connectionsInfo.GetUdpInfo()
	if mapConnectionUdp == nil {

	}
	for {
		hProcess, _, _ := procOpenProcess.Call(
			PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_VM_READ|PROCESS_ALL_ACCESS|PROCESS_QUERY_INFORMATION,
			0,
			uintptr(processEntry.Th32ProcessID),
		)
		if all == true || processEntry.Th32ProcessID == pid {
			if hProcess != 0 {
				//===================== Get ==================
				HandleProcess := syscall.Handle(hProcess)
				runTimeProcess, _ := GetRunTimeProcess(HandleProcess)
				pathProcess, _ := GetPathProcess(HandleProcess)
				commandLine, _ := GetCommandLineProcess(HandleProcess)
				KernelModuleInfos, _ := GetModuleProcess(HandleProcess)
				//KernelModuleInfos, _ := GetModuleProcess1(processEntry.Th32ProcessID)
				// ================ Token =====================
				tokenProcess, _ := GetTokenProcess(HandleProcess)
				protected, _ := GetProtectedProcess(HandleProcess)
				tokenProcess.Protected = protected

				infoProcess := ProcessInfo{
					Name:          string(processEntry.SzExeFile[:bytes.IndexByte(processEntry.SzExeFile[:], 0)]),
					Pid:           processEntry.Th32ProcessID,
					ParentPid:     processEntry.Th32ParentProcessID,
					Path:          pathProcess,
					Commandline:   commandLine,
					Runtime:       runTimeProcess,
					Token:         tokenProcess,
					Module:        KernelModuleInfos,
					ConnectionTcp: mapConnectionTcp[processEntry.Th32ProcessID],
					ConnectionUdp: mapConnectionUdp[processEntry.Th32ProcessID],
				}
				result = append(result, infoProcess)
				// ==========================================
			}
		}
		procCloseHandle.Call(hProcess)
		ret, _, err = procProcess32Next.Call(
			hProcessSnap,
			uintptr(unsafe.Pointer(&processEntry)),
		)

		if ret == 0 {
			break
		}

	}
	if len(result) == 0 {
		return nil, errors.New("Can't find process!")
	}
	return result, nil
}

func GetProtectedProcess(hProcess syscall.Handle) (string, error) {
	var size = uint32(1)
	buf := make([]byte, size)
	var sizeRt uint32
	ret, _, _ := procZwQueryInformationProcess.Call(
		uintptr(hProcess),
		ProcessProtectionInformation,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&sizeRt)),
	)
	if sizeRt == 0 {
		return "", errors.New("Error!")
	}
	ptr := (*PS_PROTECTION)(unsafe.Pointer(&buf[0]))
	if ret != 0 {
		return "", errors.New("Error!")
	}
	result := ProtectedSignerNames[ptr.Signer()]
	return result, nil
}

func GetUserNameBySID(sid *syscall.SID) (string, error) {
	var nameSize, domainSize, sidName uint32

	ret, _, err := procLookupAccountSidW.Call(
		0,
		uintptr(unsafe.Pointer(sid)),
		0,
		uintptr(unsafe.Pointer(&nameSize)),
		0,
		uintptr(unsafe.Pointer(&domainSize)),
		uintptr(unsafe.Pointer(&sidName)),
	)
	if ret == 0 && errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
		return "", nil
	}
	bufName := make([]uint16, nameSize)
	bufDomain := make([]uint16, domainSize)
	ret, _, err = procLookupAccountSidW.Call(
		0,
		uintptr(unsafe.Pointer(sid)),
		uintptr(unsafe.Pointer(&bufName[0])),
		uintptr(unsafe.Pointer(&nameSize)),
		uintptr(unsafe.Pointer(&bufDomain[0])),
		uintptr(unsafe.Pointer(&domainSize)),
		uintptr(unsafe.Pointer(&sidName)),
	)
	if ret == 0 {
		return "", err
	}
	result := syscall.UTF16ToString(bufDomain) + "\\" + syscall.UTF16ToString(bufName)
	return result, nil
}

func GetPathProcess(hProcess syscall.Handle) (string, error) {
	bufSize := uint32(50)
	for {
		buf := make([]uint16, bufSize)
		ret, _, err := procQueryFullProcessImageNameW.Call(
			uintptr(hProcess),
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&bufSize)),
		)
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == true {
			bufSize = bufSize * 2
			continue
		}
		if ret == 0 {
			return "", err
		}
		return syscall.UTF16ToString(buf), nil
	}
}
func GetRunTimeProcess(hProcess syscall.Handle) (uint32, error) {
	var creationTime FILETIME
	var exitTime FILETIME
	var userTime FILETIME
	var kernelTime FILETIME
	ret, _, err := procGetProcessTimes.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(&creationTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret == 0 {
		return 0, err
	}

	var fileTime SYSTEMTIME
	ret, _, err = procFileTimeToSystemTime.Call(
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&fileTime)),
	)
	if ret == 0 {
		return 0, err
	}
	return uint32(fileTime.wMilliseconds), nil
}

// ============================ COMMAND LINE 1 =================================
func GetCommandLineProcess(hProcess syscall.Handle) (string, error) {

	var sizeRt uint32
	size := uint32(50)
	buf := make([]uint16, size)
	for {
		buf = make([]uint16, size)
		ret, _, _ := procNtQueryInformationProcess.Call(
			uintptr(hProcess),
			ProcessCommandLineInformation,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		size = sizeRt
		if ret == STATUS_SUCCESS {
			dd := (*UNICODE_STRING)(unsafe.Pointer(&buf[0]))
			slice := unsafe.Slice(dd.Buffer, dd.Length/2)
			return string(utf16.Decode(slice)), nil
		}
		if ret != STATUS_INFO_LENGTH_MISMATCH {
			return "", nil
		}
	}
}

// ============================ COMMAND LINE 2 =================================
func GetCommandLineProcess2(hProcess syscall.Handle) (string, error) {
	pbi := windows.PROCESS_BASIC_INFORMATION{}
	var size = uint32(unsafe.Sizeof(pbi))
	var sizeRt uint32
	ret, _, err := procNtQueryInformationProcess.Call(
		uintptr(hProcess),
		ProcessBasicInformation,
		uintptr(unsafe.Pointer(&pbi)),
		uintptr(size),
		uintptr(unsafe.Pointer(&sizeRt)),
	)
	if ret != 0 {
		return "", err
	}
	//==================================================

	pebSize := uint32(unsafe.Sizeof(windows.PEB{}))
	pebBuf := make([]byte, pebSize)
	ret, _, err = procReadProcessMemory.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(pbi.PebBaseAddress)),
		uintptr(unsafe.Pointer(&pebBuf[0])),
		uintptr(pebSize),
		0,
	)
	if ret == 0 {
		return "", err
	}
	//================================================

	pebData := (*windows.PEB)(unsafe.Pointer(&pebBuf[0]))
	rtlSize := uint32(unsafe.Sizeof(windows.RTL_USER_PROCESS_PARAMETERS{}))
	rtlBuf := make([]byte, rtlSize)
	ret, _, err = procReadProcessMemory.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(pebData.ProcessParameters)),
		uintptr(unsafe.Pointer(&rtlBuf[0])),
		uintptr(rtlSize),
		0,
	)
	if ret == 0 {
		return "", err
	}

	//=============================================
	rtlData := (*windows.RTL_USER_PROCESS_PARAMETERS)(unsafe.Pointer(&rtlBuf[0]))
	cmdSize := uint32(rtlData.CommandLine.MaximumLength)
	cmdBuf := make([]uint16, cmdSize)
	ret, _, err = procReadProcessMemory.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(rtlData.CommandLine.Buffer)),
		uintptr(unsafe.Pointer(&cmdBuf[0])),
		uintptr(cmdSize),
		0,
	)
	if ret == 0 {
		return "", err
	}
	return syscall.UTF16ToString(cmdBuf[:]), nil
}

func GetModuleProcess(hProcess syscall.Handle) ([]moduleInfo, error) {
	var result []moduleInfo
	var size = uint32(32)
	var hModule = make([]uintptr, size)
	var sizeRt uint32
	ret, _, err := procEnumProcessModulesEx.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(&hModule[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&sizeRt)),
		LIST_MODULES_ALL,
	)
	if sizeRt == 0 {
		return result, errors.New("Size invalid! ")
	}
	size = sizeRt
	hModule = make([]uintptr, size)
	ret, _, err = procEnumProcessModulesEx.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(&hModule[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&sizeRt)),
		LIST_MODULES_ALL,
	)
	if ret == 0 {
		return result, err
	}

	count := sizeRt / uint32(unsafe.Sizeof(hModule[0]))
	for i := 0; i < int(count); i++ {
		var buf [MAX_PATH]uint16
		ret, _, err = procGetModuleFileNameExW.Call(
			uintptr(hProcess),
			hModule[i],
			uintptr(unsafe.Pointer(&buf[0])),
			MAX_PATH,
		)
		if ret != 0 {
			fileName := filepath.Base(syscall.UTF16ToString(buf[:ret]))
			kernelModuleInfo := moduleInfo{
				Name: fileName,
			}
			result = append(result, kernelModuleInfo)
		}
	}
	return result, nil
}

// GetModuleProcess1 =================== Get modules of process 1 =================================
func GetModuleProcess1(pid uint32) ([]moduleInfo, error) {
	var modules []moduleInfo

	snapshot, _, err := procCreateToolhelp32Snapshot.Call(
		TH32CS_SNAPMODULE|TH32CS_SNAPMODULE32,
		uintptr(pid),
	)
	if snapshot == 0 {
		return modules, err
	}
	defer procCloseHandle.Call(snapshot)

	var moduleEntry MODULEENTRY32
	moduleEntry.Size = uint32(unsafe.Sizeof(moduleEntry))

	retFirst, _, errFirst := procModule32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&moduleEntry)))
	if retFirst == 0 {
		return modules, errFirst
	}

	for {
		name := syscall.UTF16ToString(moduleEntry.szModule[:])
		modules = append(modules, moduleInfo{
			Name: name,
		})

		retNext, _, _ := procModule32NextW.Call(snapshot, uintptr(unsafe.Pointer(&moduleEntry)))
		if retNext == 0 {
			break
		}
	}

	return modules, nil
}
