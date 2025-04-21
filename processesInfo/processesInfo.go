package processesInfo

import (
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
type TokenInfo struct {
	SID       string
	SessionId uint32
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
type TOKEN_USER struct {
	SID *syscall.SID
}
type TOKEN_GROUPS struct {
	GroupCount uint32
	Groups     [1]syscall.SIDAndAttributes
}

var (
	advapi32                     = syscall.NewLazyDLL("advapi32.dll")
	kernel32                     = syscall.NewLazyDLL("Kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = kernel32.NewProc("Process32First")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procProcess32Next            = kernel32.NewProc("Process32Next")
	procOpenProcessToken         = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation      = advapi32.NewProc("GetTokenInformation")
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
				Name:      string(processEntry.szExeFile[:]),
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
		if retNext == 0 {
			break
		}
	}
	return result, nil
}
func GetTokenProcess(hProcess syscall.Handle) (TokenInfo, error) {
	var result TokenInfo
	var hToken syscall.Handle
	ret, _, err := procOpenProcessToken.Call(
		uintptr(hProcess),
		uintptr(global.TOKEN_QUERY),
		uintptr(unsafe.Pointer(&hToken)),
	)
	sid, _ := GetSIDProcess(hToken)
	sessionId, _ := GetSessionProcess(hToken)
	result.SID = sid
	result.SessionId = sessionId
	if ret == 0 {
		return result, err
	}
	return result, nil
}
func GetSIDProcess(hToken syscall.Handle) (string, error) {
	buf := make([]byte, 44)
	size := uint32(len(buf))
	ret, _, err := procGetTokenInformation.Call(
		uintptr(hToken),
		uintptr(global.TokenUser),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return "", err
	}
	tokenUser := (*TOKEN_USER)(unsafe.Pointer(&buf[0]))
	result, _ := tokenUser.SID.String()
	return result, nil
}
func GetSessionProcess(hToken syscall.Handle) (uint32, error) {
	var sessionId uint32
	var size uint32
	ret, _, err := procGetTokenInformation.Call(
		uintptr(hToken),
		uintptr(global.TokenSessionId),
		uintptr(unsafe.Pointer(&sessionId)),
		unsafe.Sizeof(sessionId),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return 0, err
	}
	return sessionId, nil
}
func GetOwnerProcess(hToken syscall.Handle) (uint32, error) {
	var sessionId uint32
	var size uint32
	ret, _, err := procGetTokenInformation.Call(
		uintptr(hToken),
		uintptr(global.TokenOwner),
		uintptr(unsafe.Pointer(&sessionId)),
		unsafe.Sizeof(sessionId),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return 0, err
	}
	return sessionId, nil
}

func getTokenGroupsSIDs(token syscall.Token) ([]string, error) {
	var size uint32
	ret, _, err := procGetTokenInformation.Call(token,
		global.TokenGroups,
		nil,
		0,
		&size,
	)
	if err != syscall.ERROR_INSUFFICIENT_BUFFER {
		return nil, err
	}

	buf := make([]byte, size)
	err = syscall.GetTokenInformation(token, syscall.TokenGroups, &buf[0], size, &size)
	if err != nil {
		return nil, err
	}

	tg := (*syscall.Tokengroups)(unsafe.Pointer(&buf[0]))
	count := int(tg.GroupCount)

	sids := []string{}
	base := uintptr(unsafe.Pointer(&tg.Groups[0]))
	sizeSIDAttr := unsafe.Sizeof(syscall.SIDAndAttributes{})

	for i := 0; i < count; i++ {
		sidAttr := (*syscall.SIDAndAttributes)(unsafe.Pointer(base + uintptr(i)*sizeSIDAttr))
		sidStr, err := convertSidToString(sidAttr.Sid)
		if err != nil {
			continue
		}
		sids = append(sids, sidStr)
	}
	return sids, nil
}
