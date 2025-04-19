package global

var (
	MAX_PATH = 260

	//================ ACESS =========================
	READ_CONTROL                      = 0x00020000
	PROCESS_ALL_ACCESS                = 0x000F0000
	PROCESS_QUERY_INFORMATION         = 0x0400
	PROCESS_VM_READ                   = 0x0010
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	TH32CS_SNAPPROCESS                = 0x00000002

	KEY_READ = 0x20019

	//================= ERROR =========================
	ERROR_MORE_DATA = uintptr(234)

	//================= REGISTRY ======================
	HKEY_LOCAL_MACHINE = 0x80000002
)
