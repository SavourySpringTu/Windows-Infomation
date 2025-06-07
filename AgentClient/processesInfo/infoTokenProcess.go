package processesInfo

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	TokenInformationClassUser      = 1
	TokenInformationClassOwner     = 4
	TokenGroups                    = 2
	TokenPrivileges                = 3
	TokenStatistics                = 10
	TokenInformationClassSessionId = 12
	TokenVirtualizationEnabled     = 24
)

const (
	TOKEN_ALL_ACCESS = 0xF01FF
	TOKEN_QUERY      = 0x0008
)

var (
	procLookupPrivilegeNameW  = advapi32.NewProc("LookupPrivilegeNameW")
	procEnumProcessModulesEx  = psapi.NewProc("EnumProcessModulesEx")
	procGetModuleFileNameExW  = psapi.NewProc("GetModuleFileNameExW")
	procOpenProcessToken      = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation   = advapi32.NewProc("GetTokenInformation")
	procLookupPrivilegeValueW = advapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges = advapi32.NewProc("AdjustTokenPrivileges")
	procGetCurrentProcess     = kernel32.NewProc("GetCurrentProcess")
)

type TOKEN_USER struct {
	User SID_AND_ATTRIBUTES
}
type TOKEN_GROUPS struct {
	GroupCount uint32
	Groups     [1]SID_AND_ATTRIBUTES
}

type SID_AND_ATTRIBUTES struct {
	Sid        *syscall.SID
	Attributes uint32
}

type TOKEN_OWNER struct {
	Owner *syscall.SID
}

type TOKEN_PRIVILEGES struct {
	PrivilegeCount uint32
	Privileges     [1]LUID_AND_ATTRIBUTES
}

type LUID_AND_ATTRIBUTES struct {
	Luid       LUID
	Attributes uint32
}

type TOKEN_STATISTICS struct {
	TokenId          LUID
	AuthenticationId LUID
}

type LUID struct {
	LowPart  uint32
	HighPart uint32
}

type TokenProcess struct {
	User         string
	SID          string
	Session      uint32
	LogonSession string
	Virtualized  string
	Protected    string
	Groups       []GroupInfo
	Privileges   []PrivilegeInfo
}

type PrivilegeInfo struct {
	LUID uint32
	Name string
}

type GroupInfo struct {
	SID  string
	Name string
}

func GetTokenProcess(hProcess syscall.Handle) (TokenProcess, error) {

	// Enable SeDebugPrivilege
	errEnable := EnableSEDebugPrivilege()
	if errEnable != nil {
		fmt.Println("EnableSeDebugPrivilege Fail!")
	}

	var result TokenProcess
	var token syscall.Token
	ret, _, err := procOpenProcessToken.Call(
		uintptr(hProcess),
		uintptr(TOKEN_ALL_ACCESS|TOKEN_QUERY),
		uintptr(unsafe.Pointer(&token)),
	)
	if ret == 0 {
		procCloseHandle.Call(uintptr(token))
		return result, err
	}
	user, _ := GetUserProcess(token)
	sid, _ := GetSIDProcess(token)
	sessionId, _ := GetSessionProcess(token)
	logonSession, _ := GetLogonSessionProcess(token)
	virtualized, _ := GetVirtualizedProcess(token)
	group, _ := GetGroupProcess(token)
	pri, _ := GetPrivilegesProcess(token)

	result.Session = sessionId
	result.User = user
	result.SID = sid
	result.Virtualized = virtualized
	result.LogonSession = logonSession
	result.Groups = group
	result.Privileges = pri
	procCloseHandle.Call(uintptr(token))
	return result, nil
}

func GetPrivilegesProcess(token syscall.Token) ([]PrivilegeInfo, error) {
	var result []PrivilegeInfo
	var size = uint32(10)
	var sizeRt uint32
	var buf []byte

	for {
		buf = make([]byte, size)
		ret, _, err := procGetTokenInformation.Call(
			uintptr(token),
			uintptr(TokenPrivileges),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 1 {
			break
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
			return result, err
		}
		size = sizeRt
	}

	tokenPri := (*TOKEN_PRIVILEGES)(unsafe.Pointer(&buf[0]))
	for _, i := range tokenPri.AllPrivileges() {
		id := i.Luid.LowPart
		name, _ := GetPrivilegeName(i.Luid)
		if name == "" {
			continue
		}
		privilege := PrivilegeInfo{
			LUID: id,
			Name: name,
		}
		result = append(result, privilege)
	}
	return result, nil
}

func GetPrivilegeName(id LUID) (string, error) {
	privilegeName := make([]uint16, 10)
	var size = uint32(len(privilegeName))
	ret, _, err := procLookupPrivilegeNameW.Call(
		0,
		uintptr(unsafe.Pointer(&id)),
		uintptr(unsafe.Pointer(&privilegeName[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	privilegeName = make([]uint16, size)
	ret, _, err = procLookupPrivilegeNameW.Call(
		0,
		uintptr(unsafe.Pointer(&id)),
		uintptr(unsafe.Pointer(&privilegeName[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return "", err
	}
	return syscall.UTF16ToString(privilegeName), nil
}

func GetGroupProcess(token syscall.Token) ([]GroupInfo, error) {
	var result []GroupInfo
	var size = uint32(10)
	var sizeRt uint32
	var buf []byte

	for {
		buf = make([]byte, size)
		ret, _, err := procGetTokenInformation.Call(
			uintptr(token),
			uintptr(TokenGroups),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 1 {
			break
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
			return result, err
		}
		size = sizeRt
	}

	tokenGroups := (*TOKEN_GROUPS)(unsafe.Pointer(&buf[0]))
	for _, i := range tokenGroups.AllGroups() {
		sidStr, _ := i.Sid.String()
		name, _ := GetUserNameBySID(i.Sid)
		groupInfo := GroupInfo{
			SID:  sidStr,
			Name: name,
		}
		result = append(result, groupInfo)
	}

	return result, nil
}

func GetVirtualizedProcess(token syscall.Token) (string, error) {
	var size = uint32(1)
	buf := make([]uint32, size)
	for {
		ret, _, err := procGetTokenInformation.Call(
			uintptr(token),
			uintptr(TokenVirtualizationEnabled),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret == 1 {
			if buf[0] == 1 {
				return "Enable", nil
			} else {
				return "Disable", nil
			}
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == true {
			buf = make([]uint32, size)
			continue
		}
	}
}
func GetLogonSessionProcess(token syscall.Token) (string, error) {
	var size = uint32(256)
	var sizeRt uint32
	var buf []byte
	for {
		buf = make([]byte, size)
		ret, _, err := procGetTokenInformation.Call(
			uintptr(token),
			uintptr(TokenStatistics),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 1 {
			tokenOrigin := (*TOKEN_STATISTICS)(unsafe.Pointer(&buf[0]))
			id := tokenOrigin.AuthenticationId
			hex := strconv.FormatUint(uint64(id.LowPart), 16)
			return hex, nil
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
			return "", err
		}
		size = sizeRt
	}
}

func GetSessionProcess(token syscall.Token) (uint32, error) {
	var sessionId uint32
	var size uint32
	ret, _, err := procGetTokenInformation.Call(
		uintptr(token),
		uintptr(TokenInformationClassSessionId),
		uintptr(unsafe.Pointer(&sessionId)),
		unsafe.Sizeof(sessionId),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return 0, err
	}
	return sessionId, nil
}
func GetSIDProcess(token syscall.Token) (string, error) {
	var sizeRt byte
	var size = byte(64)
	var buf = make([]byte, size)
	for {
		ret, _, err := procGetTokenInformation.Call(
			uintptr(token),
			uintptr(TokenInformationClassOwner),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 1 {
			tokenOwner := (*TOKEN_OWNER)(unsafe.Pointer(&buf[0]))
			result, _ := tokenOwner.Owner.String()
			return result, nil
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
			return "", err
		}
		size = sizeRt
		buf = make([]byte, size)
	}

}

func GetUserProcess(token syscall.Token) (string, error) {
	var buf = make([]uint16, 50)
	var size uint32
	var sizeRt uint32
	for {
		ret, _, err := procGetTokenInformation.Call(
			uintptr(token),
			uintptr(TokenInformationClassUser),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			uintptr(unsafe.Pointer(&size)),
		)
		size = sizeRt
		if ret != 0 {
			tokenUser := (*TOKEN_USER)(unsafe.Pointer(&buf[0]))
			result, _ := GetUserNameBySID(tokenUser.User.Sid)
			return result, nil
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
			return "", err
		}
	}

}

func EnableSEDebugPrivilege() error {
	var hToken syscall.Token
	currentProcess, _, errCurrentProcess := procGetCurrentProcess.Call()
	if currentProcess == 0 {
		return errCurrentProcess
	}
	hProcess := syscall.Handle(currentProcess)
	ret, _, err := procOpenProcessToken.Call(
		uintptr(hProcess),
		syscall.TOKEN_ADJUST_PRIVILEGES|syscall.TOKEN_QUERY,
		uintptr(unsafe.Pointer(&hToken)),
	)

	if ret == 0 {
		return err
	}

	var luid windows.LUID
	priName, _ := syscall.UTF16PtrFromString("SeDebugPrivilege")
	ret, _, err = procLookupPrivilegeValueW.Call(
		0,
		uintptr(unsafe.Pointer(priName)),
		uintptr(unsafe.Pointer(&luid)),
	)
	if ret == 0 {
		return err

	}

	privileges := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{Luid: luid, Attributes: windows.SE_PRIVILEGE_ENABLED},
		},
	}

	ret, _, err = procAdjustTokenPrivileges.Call(
		uintptr(hToken),
		0,
		uintptr(unsafe.Pointer(&privileges)),
		0,
		0,
		0,
	)
	if ret == 0 {
		return err
	}
	return nil
}
