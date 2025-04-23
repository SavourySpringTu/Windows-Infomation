package global

import (
	"syscall"
)

const (
	MAX_PATH                          = 260
	READ_CONTROL                      = 0x00020000
	PROCESS_ALL_ACCESS                = 0x000F0000
	PROCESS_QUERY_INFORMATION         = 0x0400
	PROCESS_VM_READ                   = 0x0010
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	TH32CS_SNAPPROCESS                = 0x00000002
	TOKEN_QUERY                       = 0x0008
	KEY_READ                          = 0x20019
)
const (
	ERROR_MORE_DATA uintptr = 234
)
const (
	HKEY_LOCAL_MACHINE = 0x80000002
)
const (
	FILE_ATTRIBUTE_ARCHIVE uint32 = 0x20
)

const (
	TokenUser       = 1
	TokenGroups     = 2
	TokenPrivileges = 3
	TokenOwner      = 4
	TokenStatistics = 10
	TokenSessionId  = 12
)

type SECURITY_LOGON_SESSION_DATA struct {
	Size                  uint32
	LogonId               LUID
	UserName              LSA_UNICODE_STRING
	LogonDomain           LSA_UNICODE_STRING
	AuthenticationPackage LSA_UNICODE_STRING
	LogonType             uint32
	Session               uint32
	Sid                   *syscall.SID
	LogonTime             uint64
	LogonServer           LSA_UNICODE_STRING
	DnsDomainName         LSA_UNICODE_STRING
	Upn                   LSA_UNICODE_STRING
}

type LSA_UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

type LUID struct {
	LowPart  uint32
	HighPart int32
}

// Logon types
const (
	LOGON_INTERACTIVE       = 2
	LOGON_NETWORK           = 3
	LOGON_BATCH             = 4
	LOGON_SERVICE           = 5
	LOGON_UNLOCK            = 7
	LOGON_NETWORK_CLEARTEXT = 8
	LOGON_NEW_CREDENTIALS   = 9
)
