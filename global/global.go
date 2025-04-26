package global

const (
	MAX_PATH                          = 260
	READ_CONTROL                      = 0x00020000
	PROCESS_ALL_ACCESS                = 0x000F0000
	PROCESS_QUERY_INFORMATION         = 0x0400
	PROCESS_VM_READ                   = 0x0010
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	TH32CS_SNAPPROCESS                = 0x00000002
	TOKEN_QUERY                       = 0x0008
)
const (
	ERROR_MORE_DATA uintptr = 234
)
const (
	HKEY_LOCAL_MACHINE = 0x80000002
	KEY_READ           = 0x20019
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
