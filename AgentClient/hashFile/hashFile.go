package hashFile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	MAX_PATH               = 260
	PROV_RSA_ARS           = 24
	FILE_ATTRIBUTE_ARCHIVE = 32
	CRYPT_VERIFYCONTEXT    = 0xF0000000
	HP_HASHVAL             = 0x0002
	CALG_MD5               = 0x00008003
	CALG_SHA1              = 0x8004
	CALG_SHA_256           = 0x0000800c
)

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
	cAlternateFileName [14]uint
	dwFileType         uint32
	dwCreatorType      uint32
	wFinderFlags       uint16
}
type FileHash struct {
	Name    string
	MD5     string
	SHA1    string
	SHA_256 string
}

var (
	advapi32                 = syscall.NewLazyDLL("advapi32.dll")
	kernel32                 = syscall.NewLazyDLL("Kernel32.dll")
	procCryptAcquireContextW = advapi32.NewProc("CryptAcquireContextW")
	procCryptCreateHash      = advapi32.NewProc("CryptCreateHash")
	procCryptHashData        = advapi32.NewProc("CryptHashData")
	procCryptGetHashParam    = advapi32.NewProc("CryptGetHashParam")
	procFindFirstFileW       = kernel32.NewProc("FindFirstFileW")
	procFindNextFileW        = kernel32.NewProc("FindNextFileW")
	procCryptDestroyHash     = advapi32.NewProc("CryptDestroyHash")
	procCryptReleaseContext  = advapi32.NewProc("CryptReleaseContext")
	procFindClose            = kernel32.NewProc("FindClose")
)

func GetHashFileAndFolder(path string) ([]FileHash, error) {
	var result []FileHash
	var fileAttribute WIN32_FIND_DATAA
	path = filepath.Clean(path)
	filePath, _ := syscall.UTF16PtrFromString(path)
	//==========================  check ===============================
	_, _, err := procFindFirstFileW.Call(
		uintptr(unsafe.Pointer(filePath)),
		uintptr(unsafe.Pointer(&fileAttribute)),
	)
	if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) == true {
		return result, err
	}
	//============================ IS FILE =================================
	if fileAttribute.dwFileAttributes == FILE_ATTRIBUTE_ARCHIVE {
		fileHash, e := HashFile(path)
		if e != nil {
			return result, e
		}
		fileHash.Name = syscall.UTF16ToString(fileAttribute.cFileName[:])
		result = append(result, fileHash)
		return result, nil
	}
	//============================= IS FOLDER ===============================
	pathAll := filepath.Join(path, `\*`)
	filePath, _ = syscall.UTF16PtrFromString(pathAll)
	hFile, _, err := procFindFirstFileW.Call(
		uintptr(unsafe.Pointer(filePath)),
		uintptr(unsafe.Pointer(&fileAttribute)),
	)
	defer procFindClose.Call(hFile)
	if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) == true {
		return result, err
	}
	for {
		fmt.Println(syscall.UTF16ToString(fileAttribute.cFileName[:]))
		if fileAttribute.dwFileAttributes == FILE_ATTRIBUTE_ARCHIVE {
			fullPath := path + `\` + syscall.UTF16ToString(fileAttribute.cFileName[:])
			fileHash, e := HashFile(fullPath)
			if e != nil {
				return result, e
			}
			fileHash.Name = syscall.UTF16ToString(fileAttribute.cFileName[:])
			result = append(result, fileHash)
		}
		retNext, _, errNext := procFindNextFileW.Call(
			hFile,
			uintptr(unsafe.Pointer(&fileAttribute)),
		)
		if retNext == 0 {
			if errors.Is(errNext, syscall.ERROR_NO_MORE_FILES) == true {
				break
			}
			return result, errNext
		}
	}
	return result, nil
}
func HashFile(path string) (FileHash, error) {
	var result FileHash
	calg := [3]int{
		CALG_MD5, CALG_SHA1, CALG_SHA_256,
	}
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	for i := 0; i < 3; i++ {
		// reset pointer to the start
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return result, err
		}
		var hCryptProv syscall.Handle
		ret, _, _ := procCryptAcquireContextW.Call(
			uintptr(unsafe.Pointer(&hCryptProv)),
			0, 0,
			uintptr(PROV_RSA_ARS),
			uintptr(CRYPT_VERIFYCONTEXT),
		)
		if ret == 0 {
			return result, nil
		}
		defer procCryptReleaseContext.Call(uintptr(hCryptProv))
		var hHash syscall.Handle
		ret, _, err = procCryptCreateHash.Call(
			uintptr(hCryptProv),
			uintptr(calg[i]),
			0,
			0,
			uintptr(unsafe.Pointer(&hHash)),
		)
		if ret == 0 {
			continue
		}
		defer procCryptDestroyHash.Call(uintptr(hHash))

		buf := make([]byte, 256)
		for {
			n, errRead := file.Read(buf)
			if n > 0 {
				ret, _, _ = procCryptHashData.Call(
					uintptr(hHash),
					uintptr(unsafe.Pointer(&buf[0])),
					uintptr(n),
					0,
				)
			}
			if errRead != nil {
				if errRead == io.EOF {
					break
				}
				return result, errRead
			}
		}
		hash := make([]byte, 32)
		if i == 0 {
			hash = make([]byte, 16)
		} else if i == 1 {
			hash = make([]byte, 20)
		}

		hashLen := uint32(len(hash))
		ret, _, err = procCryptGetHashParam.Call(
			uintptr(hHash),
			uintptr(HP_HASHVAL),
			uintptr(unsafe.Pointer(&hash[0])),
			uintptr(unsafe.Pointer(&hashLen)),
			0,
		)
		if ret == 0 {
			continue
		}
		hexString := fmt.Sprintf("%x", hash)
		if i == 0 {
			result.MD5 = hexString
		} else if i == 2 {
			result.SHA_256 = hexString
		} else {
			result.SHA1 = hexString
		}
	}
	return result, nil
}
