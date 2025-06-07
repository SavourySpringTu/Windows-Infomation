package filesInfo

import (
	"AgentClient/hashFile"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	MAX_PATH               = 260
	FILE_ATTRIBUTE_ARCHIVE = 32
)

var (
	kernel32                 = syscall.NewLazyDLL("Kernel32.dll")
	procGetFileAttributesExW = kernel32.NewProc("GetFileAttributesExW")
	procFileTimeToSystemTime = kernel32.NewProc("FileTimeToSystemTime")
	procFindFirstFileW       = kernel32.NewProc("FindFirstFileW")
	procFindNextFileW        = kernel32.NewProc("FindNextFileW")
	procFindClose            = kernel32.NewProc("FindClose")
)

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

type FILETIME struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}
type WIN32_FIND_DATAA struct {
	dwFileAttributes   uint32
	ftCreationTime     FILETIME
	ftLastAccessTime   FILETIME
	ftLastWriteTime    FILETIME
	nFileSizeHigh      uint32
	nFileSizeLow       uint32
	dwReserved0        uint32
	dwReserved1        uint32
	cFileName          [MAX_PATH]uint16
	cAlternateFileName [14]uint16
	dwFileType         uint32
	dwCreatorType      uint32
	wFinderFlags       uint16
}
type FileInfo struct {
	Name         string
	DateCreated  string
	DateModified string
	Size         uint64
	MD5          string
	SHA1         string
	SHA_256      string
}

func GetInfoFileAndFolder(path string) ([]FileInfo, error) {
	var result []FileInfo
	var fileAttribute WIN32_FIND_DATAA
	path = filepath.Clean(path)
	var filePath, _ = syscall.UTF16PtrFromString(path)
	fmt.Println(path)
	//==========================  check ===============================
	hFile, _, err := procFindFirstFileW.Call(
		uintptr(unsafe.Pointer(filePath)),
		uintptr(unsafe.Pointer(&fileAttribute)),
	)
	defer procFindClose.Call(hFile)
	if errors.Is(err, windows.ERROR_SUCCESS) != true {
		return result, err
	}
	//============================ IS FILE =================================
	if fileAttribute.dwFileAttributes == FILE_ATTRIBUTE_ARCHIVE {
		fileInfo, e := GetInfoFile(fileAttribute)
		if e != nil {
			return result, e
		}
		hash, _ := hashFile.HashFile(path)
		fileInfo.MD5 = hash.MD5
		fileInfo.SHA1 = hash.SHA1
		fileInfo.SHA_256 = hash.SHA_256
		result = append(result, fileInfo)
		return result, nil
	}
	//============================= IS FOLDER ===============================
	pathAll := filepath.Join(path, `\*`)
	filePath, _ = syscall.UTF16PtrFromString(pathAll)
	hFind, _, errFirst := procFindFirstFileW.Call(
		uintptr(unsafe.Pointer(filePath)),
		uintptr(unsafe.Pointer(&fileAttribute)),
	)
	defer procFindClose.Call(hFile)
	if errors.Is(errFirst, syscall.ERROR_FILE_NOT_FOUND) == true {
		return result, errFirst
	}
	for {
		if fileAttribute.dwFileAttributes == FILE_ATTRIBUTE_ARCHIVE {
			fileInfo, errGet := GetInfoFile(fileAttribute)
			if errGet != nil {
				return result, errGet
			}
			fileInfo.Name = syscall.UTF16ToString(fileAttribute.cFileName[:])
			fmt.Println(path + "//" + fileInfo.Name)
			hash, _ := hashFile.HashFile(path + `\` + fileInfo.Name)
			fileInfo.MD5 = hash.MD5
			fileInfo.SHA1 = hash.SHA1
			fileInfo.SHA_256 = hash.SHA_256
			result = append(result, fileInfo)
		}
		retNext, _, errNext := procFindNextFileW.Call(
			hFind,
			uintptr(unsafe.Pointer(&fileAttribute)),
		)
		if retNext == 0 {
			if errors.Is(errNext, syscall.ERROR_NO_MORE_FILES) == true {
				break
			}
			return result, nil
		}
	}
	return result, nil
}
func GetInfoFile(fileAttributes WIN32_FIND_DATAA) (FileInfo, error) {
	fileInfo, _ := GetTimeFile(fileAttributes)
	fileInfo.Name = syscall.UTF16ToString(fileAttributes.cFileName[:])
	fileInfo.Size = uint64(fileAttributes.nFileSizeLow)
	return fileInfo, nil
}
func GetTimeFile(fileAttributes WIN32_FIND_DATAA) (FileInfo, error) {
	var result FileInfo
	var createdTime SYSTEMTIME
	ret, _, _ := procFileTimeToSystemTime.Call(
		uintptr(unsafe.Pointer(&fileAttributes.ftCreationTime)),
		uintptr(unsafe.Pointer(&createdTime)),
	)
	var lastWriteTime SYSTEMTIME
	ret, _, _ = procFileTimeToSystemTime.Call(
		uintptr(unsafe.Pointer(&fileAttributes.ftLastWriteTime)),
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
		lastWriteTime.wMinute,
		lastWriteTime.wDay,
		lastWriteTime.wMonth,
		lastWriteTime.wYear,
	)
	return result, nil
}
