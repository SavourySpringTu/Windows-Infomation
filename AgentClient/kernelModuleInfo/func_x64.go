//go:build amd64
// +build amd64

package kernelModuleInfo

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func GetSHA256KernelModule(pathFile string) (string, error) {
	file, err := os.Open(pathFile)
	if err != nil {
		return "", err
	}
	var hCryptProv syscall.Handle
	ret, _, _ := procCryptAcquireContextW.Call(
		uintptr(unsafe.Pointer(&hCryptProv)),
		0, 0,
		uintptr(PROV_RSA_ARS),
		uintptr(CRYPT_VERIFYCONTEXT),
	)
	if ret == 0 {

		return "", err
	}
	defer procCryptReleaseContext.Call(uintptr(hCryptProv))
	var hHash syscall.Handle
	ret, _, err = procCryptCreateHash.Call(
		uintptr(hCryptProv),
		uintptr(CALG_SHA_256),
		0,
		0,
		uintptr(unsafe.Pointer(&hHash)),
	)
	if ret == 0 {
		return "", err
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
		break
		if errRead != nil {
			break
		}
	}
	hash := make([]byte, 32)

	hashLen := uint32(len(hash))
	ret, _, err = procCryptGetHashParam.Call(
		uintptr(hHash),
		uintptr(HP_HASHVAL),
		uintptr(unsafe.Pointer(&hash[0])),
		uintptr(unsafe.Pointer(&hashLen)),
		0,
	)
	if ret == 0 {
		return "", err
	}
	hexString := fmt.Sprintf("%x", hash)
	return hexString, nil
}
