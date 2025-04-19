package filesInfo

import (
	"syscall"
	"unsafe"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procFindFirstFileW     = kernel32.NewProc("FindFirstFileW")
	procFindNextFileW      = kernel32.NewProc("FindNextFileW")
	procFindClose          = kernel32.NewProc("FindClose")
	procGetFileAttributesW = kernel32.NewProc("GetFileAttributesW")

	FILE_ATTRIBUTE_DIRECTORY uint32 = 0x10
)

type FILETIME struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}
type WIN32_FILE_ATTRIBUTE_DATA struct {
	dwFileAttributes uint32
	ftCreationTime   FILETIME
	ftLastAccessTime FILETIME
	ftLastWriteTime  FILETIME
	nFileSizeHigh    uint32
	nFileSizeLow     uint32
}

type FileInfo struct {
	Name         string
	DateCreated  string
	DateModified string
	Size         int64
}

func GetInfoFileAndFolder(path string) {
	var fileAttributes WIN32_FILE_ATTRIBUTE_DATA
	pathPtr, _ := syscall.UTF16PtrFromString(path)
	ret, _, e := procGetFileAttributesW.Call(
		uintptr(unsafe.Pointer(&pathPtr)),
		0,
		uintptr(unsafe.Pointer(&fileAttributes)),
	)
	if ret == 0 {
		return
	}
}
