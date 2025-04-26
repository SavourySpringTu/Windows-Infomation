package processesInfo

import (
	"bytes"
	"main/global"
	"syscall"
	"unsafe"
)

type ProcessInfo struct {
	Pid         uint32
	PidParent   uint32
	Runtime     uint32
	Name        string
	Path        string
	CommandLine string
	Token       TokenInfo
}

type PROCESSENTRY32 struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]byte
}

var (
	advapi32                     = syscall.NewLazyDLL("advapi32.dll")
	kernel32                     = syscall.NewLazyDLL("Kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = kernel32.NewProc("Process32First")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procProcess32Next            = kernel32.NewProc("Process32Next")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
)

func GetProcessesInfo() ([]ProcessInfo, error) {
	var result []ProcessInfo
	var processEntry PROCESSENTRY32

	hProcessSnap, _, e := procCreateToolhelp32Snapshot.Call(
		uintptr(global.TH32CS_SNAPPROCESS),
		0,
	)
	defer procCloseHandle.Call(hProcessSnap)
	if hProcessSnap == 0 {
		return result, e
	}
	processEntry.dwSize = uint32(unsafe.Sizeof(processEntry))
	ret, _, errFirst := procProcess32First.Call(
		hProcessSnap,
		uintptr(unsafe.Pointer(&processEntry)),
	)
	if ret == 0 {
		return result, errFirst
	}
	for {
		hProcess, _, _ := procOpenProcess.Call(
			uintptr(global.PROCESS_VM_READ|global.PROCESS_QUERY_LIMITED_INFORMATION),
			0,
			uintptr(processEntry.th32ProcessID),
		)
		if hProcess != 0 {
			tokenInfo, _ := GetTokenProcess(syscall.Handle(hProcess))
			var processInfo = ProcessInfo{
				Name:      string(processEntry.szExeFile[:bytes.IndexByte(processEntry.szExeFile[:], 0)]),
				Pid:       processEntry.th32ProcessID,
				PidParent: processEntry.th32ParentProcessID,
				Token:     tokenInfo,
			}
			result = append(result, processInfo)
		}
		retNext, _, _ := procProcess32Next.Call(
			hProcessSnap,
			uintptr(unsafe.Pointer(&processEntry)),
		)
		procCloseHandle.Call(hProcess)
		if retNext == 0 {
			break
		}
	}
	return result, nil
}
