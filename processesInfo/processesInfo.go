package processesInfo

import (
	"syscall"
	"unsafe"
)

const (
	READ_CONTROL              = 0x00020000
	PROCESS_ALL_ACCESS        = 0x000F0000
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

type ProcessInfo struct {
	Pid       uint32
	PidParent uint32
	Runtime   uint32
	Name      string
	Path      string
}

type PROCESS_BASIC_INFORMATION struct {
	Reserved1                    uintptr
	PebBaseAddress               uintptr
	Reserved2                    [2]uintptr
	UniqueProcessId              uintptr
	InheritedFromUniqueProcessId uintptr // <- đây là Parent PID
}

type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        uintptr
}

type PEB struct {
	_                 [2]byte
	ProcessParameters uintptr
}
type RTL_USER_PROCESS_PARAMETERS struct {
	_           [16]byte
	CommandLine UNICODE_STRING
}

var (
	psapi                         = syscall.NewLazyDLL("Psapi.dll")
	kernel32                      = syscall.NewLazyDLL("Kernel32.dll")
	ntdll                         = syscall.NewLazyDLL("ntdll.dll")
	procEnumProcesses             = psapi.NewProc("EnumProcesses")
	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procGetModuleBaseNameW        = psapi.NewProc("GetModuleBaseNameW")
	procNtQueryInformationProcess = ntdll.NewProc("NtQueryInformationProcess")
	procReadProcessMemory         = kernel32.NewProc("ReadProcessMemory")
)

func GetProcessesInfo() ([]ProcessInfo, error) {
	var result []ProcessInfo
	var buf = make([]uint32, 1)
	var size uint32

	ret, _, err := procEnumProcesses.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return result, err
	}
	buf = make([]uint32, size/4)
	ret, _, err = procEnumProcesses.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return result, err
	}
	for _, i := range buf {
		GetProcessInfo, err := GetProcessInfo(i)
		if err != nil {
			continue
		}
		result = append(result, GetProcessInfo)
	}

	return result, nil
}

func GetProcessInfo(pid uint32) (ProcessInfo, error) {
	var result ProcessInfo
	ret, _, err := procOpenProcess.Call(
		uintptr(PROCESS_VM_READ|PROCESS_QUERY_INFORMATION),
		0,
		uintptr(pid),
	)
	if ret == 0 {
		return result, err
	}
	hProcess := syscall.Handle(ret)
	defer syscall.CloseHandle(hProcess)

	result.Pid = pid
	nameProcess, err := GetNameProcess(hProcess, pid)
	result.Name = nameProcess
	pidParent, err := GetParentIdProcess(hProcess, pid)
	result.PidParent = pidParent
	return result, nil
}

func GetNameProcess(hProcess syscall.Handle, pid uint32) (string, error) {
	buf := make([]uint16, 260)
	ret, _, err := procGetModuleBaseNameW.Call(
		uintptr(hProcess),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 {
		return "", err
	}
	return syscall.UTF16ToString(buf), nil
}

func GetParentIdProcess(hProcess syscall.Handle, pid uint32) (uint32, error) {
	var pbi PROCESS_BASIC_INFORMATION
	var size uint32
	ret, _, err := procNtQueryInformationProcess.Call(
		uintptr(hProcess),
		0,
		uintptr(unsafe.Pointer(&pbi)),
		unsafe.Sizeof(pbi),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 {
		return 0, err
	}
	// var peb PEB
	// var read uint32

	// ret, _, err = procReadProcessMemory.Call(
	// 	uintptr(hProcess),
	// 	pbi.PebBaseAddress,
	// 	uintptr(unsafe.Pointer(&peb)),
	// 	unsafe.Sizeof(peb),
	// 	uintptr(unsafe.Pointer(&read)),
	// )
	// if ret == 0 {
	// 	return 0, err
	// }

	// var params RTL_USER_PROCESS_PARAMETERS
	// ret, _, err = procReadProcessMemory.Call(
	// 	uintptr(hProcess),
	// 	uintptr(peb.ProcessParameters),
	// 	uintptr(unsafe.Pointer(&params)),
	// 	unsafe.Sizeof(params),
	// 	uintptr(unsafe.Pointer(&read)),
	// )
	// if ret == 0 {
	// 	fmt.Println(err)
	// }

	// buf := make([]uint16, 256)

	// ret, _, err = procReadProcessMemory.Call(
	// 	uintptr(hProcess),
	// 	uintptr(params.CommandLine.Buffer),
	// 	uintptr(unsafe.Pointer(&buf[0])),
	// 	uintptr(params.CommandLine.Length),
	// 	uintptr(unsafe.Pointer(&read)),
	// )
	// if ret == 0 {
	// 	fmt.Println(err)
	// }

	// fmt.Println("Command Line:", syscall.UTF16ToString(buf))

	return uint32(pbi.InheritedFromUniqueProcessId), nil
}
