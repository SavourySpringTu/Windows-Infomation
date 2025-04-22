package processesInfo

import (
	"bytes"
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
	Token       TokenInfo
}
type TokenInfo struct {
	SID       string
	SessionId uint32
	Group     []Group
}
type Group struct {
	SID string
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
	Groups     uintptr
}
type SID_AND_ATTRIBUTES struct {
	SID        *syscall.SID
	Attributes uint32
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
		if retNext == 0 {
			break
		}
	}
	return result, nil
}
func GetTokenProcess(hProcess syscall.Handle) (TokenInfo, error) {
	var result TokenInfo
	var hToken syscall.Token
	ret, _, err := procOpenProcessToken.Call(
		uintptr(hProcess),
		uintptr(global.TOKEN_QUERY),
		uintptr(unsafe.Pointer(&hToken)),
	)
	sid, _ := GetSIDProcess(hToken)
	sessionId, _ := GetSessionProcess(hToken)
	group, _ := GetGroupProcess(hToken)
	result.SID = sid
	result.SessionId = sessionId
	result.Group = group
	if ret == 0 {
		return result, err
	}
	return result, nil
}
func GetSIDProcess(hToken syscall.Token) (string, error) {
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
func GetSessionProcess(hToken syscall.Token) (uint32, error) {
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

func GetGroupProcess(token syscall.Token) ([]Group, error) {
	var result []Group
	buf := make([]byte, 10)
	var size = uint32(len(buf))
	ret, _, err := procGetTokenInformation.Call(
		uintptr(token),
		uintptr(global.TokenGroups),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		fmt.Println(err)
	}
	buf = make([]byte, size)
	ret, _, err = procGetTokenInformation.Call(
		uintptr(token),
		uintptr(global.TokenGroups),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&size)),
	)
	tokenGroups := (*TOKEN_GROUPS)(unsafe.Pointer(&buf[0]))
	fmt.Println("addrtokenGroups:", uintptr(unsafe.Pointer(tokenGroups)))
	fmt.Println("sizeofCount:", unsafe.Sizeof(tokenGroups.GroupCount))
	fmt.Println("tokenGroups.group:", tokenGroups.Groups)
	groupCount := tokenGroups.GroupCount
	groupsPtr := uintptr(unsafe.Pointer(&tokenGroups)) + unsafe.Sizeof(tokenGroups.GroupCount)
	fmt.Println("groupsPtr:", groupsPtr)

	for i := 0; i < int(groupCount); i++ {
		sidAttrAddr := groupsPtr + +uintptr(i)*(unsafe.Sizeof(SID_AND_ATTRIBUTES{}))
		sidAttr := (*SID_AND_ATTRIBUTES)(unsafe.Pointer(sidAttrAddr))
		sidAddr := sidAttr.SID
		sid := (*syscall.SID)(unsafe.Pointer(&sidAddr))
		strSid, _ := sid.String()
		fmt.Println(strSid)
	}

	return result, nil
}
