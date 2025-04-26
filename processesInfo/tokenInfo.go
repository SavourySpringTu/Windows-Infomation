package processesInfo

import (
	"errors"
	"main/global"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	procOpenProcessToken     = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation  = advapi32.NewProc("GetTokenInformation")
	procLookupPrivilegeNameW = advapi32.NewProc("LookupPrivilegeNameW")
)

type TOKEN_USER struct {
	SID *syscall.SID
}
type TOKEN_GROUPS struct {
	GroupCount uint32
	Groups     [1]SID_AND_ATTRIBUTES
}
type SID_AND_ATTRIBUTES struct {
	SID        *syscall.SID
	Attributes uint32
}
type LUID struct {
	LowPart  uint32
	HighPart uint32
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

type GroupInfo struct {
	SID string
}
type PrivilegeInfo struct {
	Name string
}
type TokenInfo struct {
	SID          string
	SessionId    uint32
	LogonSession string
	Groups       []GroupInfo
	Privileges   []PrivilegeInfo
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
	logon, _ := GetLogonSessionInfo(hToken)
	pri, _ := GetPrivilegesProcess(hToken)
	result.SID = sid
	result.SessionId = sessionId
	result.LogonSession = logon
	result.Groups = group
	result.Privileges = pri
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
func (g *TOKEN_GROUPS) AllGroups() []SID_AND_ATTRIBUTES {
	return (*[(1 << 28) - 1]SID_AND_ATTRIBUTES)(unsafe.Pointer(&g.Groups[0]))[:g.GroupCount:g.GroupCount]
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
			uintptr(global.TokenGroups),
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
		sidStr, _ := i.SID.String()
		group := GroupInfo{
			SID: sidStr,
		}
		result = append(result, group)
	}
	return result, nil
}
func (p *TOKEN_PRIVILEGES) AllPrivileges() []LUID_AND_ATTRIBUTES {
	return (*[(1 << 27) - 1]LUID_AND_ATTRIBUTES)(unsafe.Pointer(&p.Privileges[0]))[:p.PrivilegeCount:p.PrivilegeCount]
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
			uintptr(global.TokenPrivileges),
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
		name, _ := GetNamePrivilege(i.Luid)
		privilege := PrivilegeInfo{
			Name: name,
		}
		result = append(result, privilege)
	}
	return result, nil
}

func GetNamePrivilege(id LUID) (string, error) {
	var buf []byte
	size := uint32(4)
	for {
		buf = make([]byte, size)
		ret, _, err := procLookupPrivilegeNameW.Call(
			0,
			uintptr(unsafe.Pointer(&id)),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret == 1 {
			return string(buf[:size]), err
		}
		if errors.Is(err, syscall.ERROR_INSUFFICIENT_BUFFER) == false {
			return "", err
		}
	}
}

func GetLogonSessionInfo(token syscall.Token) (string, error) {
	var size = uint32(10)
	var sizeRt uint32
	var buf []byte
	for {
		buf = make([]byte, size)
		ret, _, err := procGetTokenInformation.Call(
			uintptr(token),
			uintptr(global.TokenStatistics),
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
