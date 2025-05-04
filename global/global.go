package global

const (
	READ_CONTROL                      = 0x00020000
	PROCESS_ALL_ACCESS                = 0x000F0000
	PROCESS_QUERY_INFORMATION         = 0x0400
	PROCESS_VM_READ                   = 0x0010
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	TH32CS_SNAPPROCESS                = 0x00000002
	TOKEN_QUERY                       = 0x0008
	PROCESS_DUP_HANDLE                = 0x00000040
	DUPLICATE_SAME_ACCESS             = 0x00000002
)
const (
	STATUS_SUCCESS                      = 0x00000000
	STATUS_INFO_LENGTH_MISMATCH         = 0xC0000004
	ERROR_MORE_DATA             uintptr = 234
	ERROR_INSUFFICIENT_BUFFER           = 122
	ERROR_NO_MORE_FILES                 = 18
)
const (
	UDP_TABLE_OWNER_PID = 1
	HKEY_LOCAL_MACHINE  = 0x80000002
	KEY_READ            = 0x20019
	AF_INET             = 2
	AF_INET6            = 23
)
const (
	FILE_ATTRIBUTE_ARCHIVE uint32 = 0x20
)

const (
	TokenUser               = 1
	TokenGroups             = 2
	TokenPrivileges         = 3
	TokenOwner              = 4
	TokenStatistics         = 10
	TokenSessionId          = 12
	ObjectNameInformation   = 1
	SystemHandleInformation = 16
)
