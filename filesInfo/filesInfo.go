package filesInfo

import (
	"errors"
	"fmt"
	"main/global"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procFindFirstFileW       = kernel32.NewProc("FindFirstFileW")
	procFileTimeToSystemTime = kernel32.NewProc("FileTimeToSystemTime")
	procFindNextFileW        = kernel32.NewProc("FindNextFileW")
	procFindClose            = kernel32.NewProc("FindClose")
	procGetFileAttributesExW = kernel32.NewProc("GetFileAttributesExW")
)

type FILETIME struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}
type SYSTEMTIME struct {
	wYear         uint16
	wMonth        uint16
	wDayOfWeek    uint16
	wDay          uint16
	wHour         uint16
	wMinute       uint16
	wSecond       uint16
	wMilliseconds uint16
}
type WIN32_FILE_ATTRIBUTE_DATA struct {
	dwFileAttributes uint32
	ftCreationTime   FILETIME
	ftLastAccessTime FILETIME
	ftLastWriteTime  FILETIME
	nFileSizeHigh    uint32
	nFileSizeLow     uint32
}
type WIN32_FIND_DATA struct {
	FileAttributes    uint32
	CreationTime      FILETIME
	LastAccessTime    FILETIME
	LastWriteTime     FILETIME
	FileSizeHigh      uint32
	FileSizeLow       uint32
	Reserved0         uint32
	Reserved1         uint32
	FileName          [260]uint16
	AlternateFileName [14]uint16
}
type FileInfo struct {
	Name         string
	DateCreated  string
	DateModified string
	Size         int64
}

func GetInfoFilesAndFolder(path string) ([]FileInfo, error) {
	var result []FileInfo
	var fileAttributes WIN32_FIND_DATA
	path = filepath.Clean(path) // clean
	pathPtr, _ := syscall.UTF16PtrFromString(path)
	ret, _, e := procGetFileAttributesExW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(unsafe.Pointer(&fileAttributes)),
	)
	if ret == 0 {
		return result, e
	}
	//================= IS FILE ==========================
	if fileAttributes.FileAttributes == global.FILE_ATTRIBUTE_ARCHIVE {

		fileInfo, eFileInfo := GetInfoFile(fileAttributes)
		if eFileInfo != nil {
			return result, eFileInfo
		}
		result = append(result, fileInfo)
		return result, nil
	}
	//================= IS FOLDER ========================
	pathAll := filepath.Join(path, "*")
	pathPtr, _ = syscall.UTF16PtrFromString(pathAll)
	hFile, _, eFirst := procFindFirstFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&fileAttributes)),
	)
	defer procFindClose.Call(hFile)
	if hFile == 0 {
		return result, eFirst
	}
	for {
		name := syscall.UTF16ToString(fileAttributes.FileName[:])
		if name != "." && name != ".." {
			fileInfo, eFileInfo := GetInfoFile(fileAttributes)
			if eFileInfo == nil {
				result = append(result, fileInfo)
			}
		}
		ret, _, e = procFindNextFileW.Call(
			hFile,
			uintptr(unsafe.Pointer(&fileAttributes)),
		)
		if errors.Is(e, syscall.ERROR_NO_MORE_FILES) == true {
			break
		}
	}
	return result, nil
}
func GetInfoFile(fileAttributes WIN32_FIND_DATA) (FileInfo, error) {
	fileInfo, _ := GetTimeFile(fileAttributes)
	fileInfo.Name = syscall.UTF16ToString(fileAttributes.FileName[:])
	fileInfo.Size = (int64(fileAttributes.FileSizeHigh) << 32) + int64(fileAttributes.FileSizeLow)
	return fileInfo, nil
}
func GetTimeFile(fileAttributes WIN32_FIND_DATA) (FileInfo, error) {
	var result FileInfo
	var createdTime SYSTEMTIME
	ret, _, _ := procFileTimeToSystemTime.Call(
		uintptr(unsafe.Pointer(&fileAttributes.CreationTime)),
		uintptr(unsafe.Pointer(&createdTime)),
	)
	var lastWriteTime SYSTEMTIME
	ret, _, _ = procFileTimeToSystemTime.Call(
		uintptr(unsafe.Pointer(&fileAttributes.LastWriteTime)),
		uintptr(unsafe.Pointer(&lastWriteTime)),
	)
	if ret == 0 {

	}
	result.DateCreated = fmt.Sprintf("%02d:%02d %02d/%02d/%04d",
		createdTime.wHour,
		createdTime.wMinute,
		createdTime.wDay,
		createdTime.wMonth,
		createdTime.wYear,
	)

	result.DateModified = fmt.Sprintf("%02d:%02d %02d/%02d/%04d",
		lastWriteTime.wHour,
		lastWriteTime.wMinute, lastWriteTime.wDay,
		lastWriteTime.wMonth,
		lastWriteTime.wYear,
	)
	return result, nil
}
