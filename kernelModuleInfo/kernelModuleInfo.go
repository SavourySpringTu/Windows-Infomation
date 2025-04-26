package kernelModuleInfo

import (
	"fmt"
	"main/global"
	"syscall"
	"unsafe"
)

var (
	advapi            = syscall.NewLazyDLL("Advapi32.dll")
	procRegOpenKeyExW = advapi.NewProc("RegOpenKeyExW")
	procRegEnumKeyExW = advapi.NewProc("RegEnumKeyExW")
)

const (
	pathKey = `SYSTEM\CurrentControlSet\Services`
)

type KernelModuleInfo struct {
	Name    string
	Version string
	Path    string
	Status  string
}

func GetKernelModuleInfo() ([]KernelModuleInfo, error) {
	var result []KernelModuleInfo
	var hKey syscall.Handle
	pathKeyPtr, _ := syscall.UTF16PtrFromString(pathKey)
	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(global.HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(pathKeyPtr)),
		0,
		uintptr(global.KEY_READ),
		uintptr(unsafe.Pointer(&hKey)),
	)
	if ret != 0 {
		return result, err
	}
	var size = uint32(20)
	var name = make([]byte, size)
	var sizeRt uint32
	fmt.Println("toi dasy")
	for i := 0; ; i++ {
		ret, _, err = procRegEnumKeyExW.Call(
			uintptr(hKey),
			uintptr(i),
			uintptr(unsafe.Pointer(&name[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
			0, 0, 0, 0,
		)

	}
}
