package processesInfo

import (
	"fmt"
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
	user32                       = syscall.NewLazyDLL("user32.dll")
	kernel32                     = syscall.NewLazyDLL("Kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = kernel32.NewProc("Process32First")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procProcess32Next            = kernel32.NewProc("Process32Next")
	procReadProcessMemory        = kernel32.NewProc("ReadProcessMemory")
	procEnumProcessModules       = kernel32.NewProc("EnumProcessModules")
	procGetModuleBaseNameW       = kernel32.NewProc("GetModuleBaseNameW")
	procGetCommandLineW          = kernel32.NewProc("GetCommandLineW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	maxPath                      = 260
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
	ret, _, e := procProcess32First.Call(
		hProcessSnap,
		uintptr(unsafe.Pointer(&processEntry)),
	)
	if ret == 0 {
		return result, e
	}
	for {
		hProcess, _, _ := procOpenProcess.Call(
			uintptr(global.PROCESS_VM_READ|global.PROCESS_QUERY_LIMITED_INFORMATION),
			0,
			uintptr(processEntry.th32ProcessID),
		)
		if hProcess != 0 {
			hProcess1 := syscall.Handle(hProcess)
			processComandLine, _ := GetProcessCommandLine(hProcess1, processEntry.th32ProcessID)
			var processInfo = ProcessInfo{
				Name:        string(processEntry.szExeFile[:]),
				Pid:         processEntry.th32ProcessID,
				PidParent:   processEntry.th32ParentProcessID,
				CommandLine: processComandLine,
			}
			result = append(result, processInfo)
		}
		ret, _, _ := procProcess32Next.Call(
			uintptr(hProcessSnap),
			uintptr(unsafe.Pointer(&processEntry)),
		)
		if ret == 0 {
			break
		}
	}
	return result, nil
}

func GetProcessCommandLine(hProcess syscall.Handle, pid uint32) (string, error) {
	var pebBaseAddress uintptr
	var read uint32
	procReadProcessMemory.Call(
		uintptr(hProcess),
		0x7ffdf000,
		uintptr(unsafe.Pointer(&pebBaseAddress)),
		unsafe.Sizeof(pebBaseAddress),
		uintptr(unsafe.Pointer(&read)),
	)

	var processParameters uintptr
	procReadProcessMemory.Call(
		uintptr(hProcess),
		pebBaseAddress+0x10,
		uintptr(unsafe.Pointer(&processParameters)),
		unsafe.Sizeof(processParameters),
		uintptr(unsafe.Pointer(&read)),
	)

	var commandLineBuffer [260]byte
	procReadProcessMemory.Call(
		uintptr(hProcess),
		processParameters+0x60,
		uintptr(unsafe.Pointer(&commandLineBuffer)),
		260,
		uintptr(unsafe.Pointer(&read)),
	)

	return string(commandLineBuffer[:]), nil
}

func GetProcessCommandLine1(pid uint32) string {
	hProcess, _, _ := procOpenProcess.Call(
		uintptr(global.PROCESS_VM_READ|global.PROCESS_QUERY_LIMITED_INFORMATION),
		0,
		uintptr(pid),
	)
	if hProcess == 0 {
		return ""
	}
	defer procCloseHandle.Call(hProcess)
	var pebBaseAddress uintptr
	var read uint32
	ret, _, _ := procReadProcessMemory.Call(
		uintptr(hProcess),
		0x7ffdf000,
		uintptr(unsafe.Pointer(&pebBaseAddress)),
		unsafe.Sizeof(pebBaseAddress),
		uintptr(unsafe.Pointer(&read)),
	)
	fmt.Println("ret", ret)
	var processParameters uintptr
	ret1, _, _ := procReadProcessMemory.Call(
		uintptr(hProcess),
		pebBaseAddress+0x10,
		uintptr(unsafe.Pointer(&processParameters)),
		unsafe.Sizeof(processParameters),
		uintptr(unsafe.Pointer(&read)),
	)
	fmt.Println("ret1", ret1)
	var commandLineBuffer [260]byte
	ret2, _, _ := procReadProcessMemory.Call(
		uintptr(hProcess),
		processParameters+0x60,
		uintptr(unsafe.Pointer(&commandLineBuffer)),
		260,
		uintptr(unsafe.Pointer(&read)),
	)
	fmt.Println("ret2", ret2)
	return string(commandLineBuffer[:])
}
