package filesinfo

import (
	"fmt"
	"path/filepath"
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
	MAX_PATH                        = 260
)

type win32FindData struct {
	FileAttributes    uint32
	_                 [3]uint32
	CreationTime      [2]uint32
	LastAccessTime    [2]uint32
	LastWriteTime     [2]uint32
	FileSizeHigh      uint32
	FileSizeLow       uint32
	Reserved0         uint32
	Reserved1         uint32
	FileName          [260]uint16
	AlternateFileName [14]uint16
}

func utf16PtrFromString(s string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(s)
	return ptr
}

func isDirectory(path string) (bool, error) {
	attrRaw, _, err := procGetFileAttributesW.Call(uintptr(unsafe.Pointer(utf16PtrFromString(path))))
	if attrRaw == 0xFFFFFFFF {
		return false, err
	}
	attr := uint32(attrRaw)
	return attr&FILE_ATTRIBUTE_DIRECTORY != 0, nil
}

// Giả sử đây là hàm xử lý file
func Readfile(path string) {
	fmt.Println("readfile:", path)
}

func ProcessPath(path string) error {
	isDir, err := isDirectory(path)
	if err != nil {
		return fmt.Errorf("cannot check file attributes: %w", err)
	}

	if !isDir {
		Readfile(path)
		return nil
	}

	searchPattern := filepath.Join(path, "*")
	var findData win32FindData

	handle, _, _ := procFindFirstFileW.Call(
		uintptr(unsafe.Pointer(utf16PtrFromString(searchPattern))),
		uintptr(unsafe.Pointer(&findData)),
	)
	if handle == uintptr(syscall.InvalidHandle) {
		return fmt.Errorf("FindFirstFileW failed")
	}
	defer procFindClose.Call(handle)

	for {
		filename := syscall.UTF16ToString(findData.FileName[:])
		if filename != "." && filename != ".." {
			fullPath := filepath.Join(path, filename)
			if findData.FileAttributes&FILE_ATTRIBUTE_DIRECTORY == 0 {
				Readfile(fullPath)
			}
		}
		ret, _, _ := procFindNextFileW.Call(handle, uintptr(unsafe.Pointer(&findData)))
		if ret == 0 {
			break
		}
	}

	return nil
}
