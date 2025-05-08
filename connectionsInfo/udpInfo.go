package connectionsInfo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"main/global"
	"net"
	"syscall"
	"unsafe"
)

var (
	iphlpapi = syscall.NewLazyDLL("iphlpapi.dll")
	ws2_32   = syscall.NewLazyDLL("ws2_32.dll")
	ntdll    = syscall.NewLazyDLL("ntdll.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	procntohl                    = ws2_32.NewProc("ntohl")
	procOpenProcessToken         = advapi32.NewProc("OpenProcessToken")
	procLookupPrivilegeValueW    = advapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges    = advapi32.NewProc("AdjustTokenPrivileges")
	procGetExtendedUdpTable      = iphlpapi.NewProc("GetExtendedUdpTable")
	procGetSockName              = ws2_32.NewProc("getsockname")
	procCloseSocket              = ws2_32.NewProc("closesocket")
	procNtDuplicateObject        = ntdll.NewProc("NtDuplicateObject")
	procNtQuerySystemInformation = ntdll.NewProc("NtQuerySystemInformation")
	procNtQueryObject            = ntdll.NewProc("NtQueryObject")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procGetpeername              = ws2_32.NewProc("getpeername")
	procWSAGetLastErr            = ws2_32.NewProc("WSAGetLastError")
	procGetCurrentProcess        = kernel32.NewProc("GetCurrentProcess")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
)

//================= SeDebug =============================

type LUID struct {
	LowPart  uint32
	HighPart int32
}

type LUIDAndAttributes struct {
	Luid       LUID
	Attributes uint32
}

type TokenPrivileges struct {
	PrivilegeCount uint32
	Privileges     [1]LUIDAndAttributes
}

//=========================================================

type SYSTEM_HANDLE struct {
	ProcessId       uint32
	ObjectTypeIndex byte
	Flags           byte
	HandleValue     uint16
	Object          uintptr
	GrantedAccess   uint32
}

type SYSTEM_HANDLE_INFORMATION struct {
	NumberOfHandles uint32
	Handles         [1]SYSTEM_HANDLE
}

type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

type MIB_UDPTABLE_OWNER_PID struct {
	dwNumEntries uint32
	UdpRow       [1]MIB_UDPROW_OWNER_PID
}

type MIB_UDPROW_OWNER_PID struct {
	dwLocalAddr uint32
	dwLocalPort uint32
	dwOwningPid uint32
}

type OBJECT_NAME_INFORMATION struct {
	Name UNICODE_STRING
}

type UPDInfo struct {
	PID        uint32
	Handle     uint16
	RemoteAddr net.IP
	LocalAddr  net.IP
	LocalPort  uint16
	RemotePort uint16
}

type Sockaddr struct {
	Family  uint16
	Port    uint16
	Address InAddr
	Zero    [8]byte
}

type InAddr struct {
	Addr [4]byte
}

func GetUdpInfo() (map[uint32][]UPDInfo, error) {
	err := EnableSeDebugPrivilege()
	if err != nil {
		fmt.Println("Enable SeDebugPrivilege fail:", err)
	} else {
		fmt.Println("Enable SeDebugPrivilege Success.")
	}

	//========================================================
	allHandle, result, errAllHandle := FilterHandle()
	if errAllHandle != nil {
		return result, errAllHandle
	}

	hCurrent, _, errCurrent := procGetCurrentProcess.Call()
	if hCurrent == 0 {
		return result, errCurrent
	}
	for _, handle := range allHandle {
		dupHandle, errDup := DupicateHandle(handle.PID, syscall.Handle(handle.Handle), syscall.Handle(hCurrent))
		if errDup != nil || isSocket(dupHandle) != true {
			continue
		}
		localAddr, localPort, remoteAddr, remotePort, _ := GetSocketAddresses(dupHandle)
		udpInfo := UPDInfo{
			PID:        handle.PID,
			Handle:     handle.Handle,
			LocalAddr:  localAddr,
			RemoteAddr: remoteAddr,
			LocalPort:  localPort,
			RemotePort: remotePort,
		}
		result[handle.PID] = append(result[handle.PID], udpInfo)
	}
	procCloseHandle.Call(hCurrent)
	return result, nil
}

func FilterHandle() ([]UPDInfo, map[uint32][]UPDInfo, error) {
	var result []UPDInfo
	allUdp, errUdp := GetAllUdp()
	if errUdp != nil {
		return result, allUdp, errUdp
	}
	allHandle, errHandle := GetAllHandle()
	if errHandle != nil {
		return result, allUdp, errHandle
	}
	for _, handle := range allHandle {
		_, exists := allUdp[handle.ProcessId]
		if exists {
			udpInfo := UPDInfo{
				PID:    handle.ProcessId,
				Handle: handle.HandleValue,
			}
			result = append(result, udpInfo)
		}
	}
	return result, allUdp, nil
}

func (g *SYSTEM_HANDLE_INFORMATION) AllHandle() []SYSTEM_HANDLE {
	return (*[(1 << 28) - 1]SYSTEM_HANDLE)(unsafe.Pointer(&g.Handles[0]))[:g.NumberOfHandles:g.NumberOfHandles]
}

func GetAllHandle() ([]SYSTEM_HANDLE, error) {

	var result []SYSTEM_HANDLE
	var size = uint32(32)
	var buf = make([]byte, size)
	var sizeRt uint32
	for {
		ret, _, _ := procNtQuerySystemInformation.Call(
			global.SystemHandleInformation,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 0 {
			break
		}
		if ret != global.STATUS_INFO_LENGTH_MISMATCH {
			return result, syscall.Errno(ret)
		}
		size = sizeRt
		buf = make([]byte, size)
	}
	ptr := (*SYSTEM_HANDLE_INFORMATION)(unsafe.Pointer(&buf[0]))
	result = ptr.AllHandle()
	return result, nil
}

func (g *MIB_UDPTABLE_OWNER_PID) AllUdp() []MIB_UDPROW_OWNER_PID {
	return (*[(1 << 28) - 1]MIB_UDPROW_OWNER_PID)(unsafe.Pointer(&g.UdpRow[0]))[:g.dwNumEntries:g.dwNumEntries]
}

func GetAllUdp() (map[uint32][]UPDInfo, error) {
	var result = make(map[uint32][]UPDInfo)
	var size = uint32(32)
	var buf = make([]byte, size)
	for {
		ret, _, _ := procGetExtendedUdpTable.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
			global.AF_INET,
			global.UDP_TABLE_OWNER_PID,
			0,
		)
		if ret == 0 {
			break
		}
		if ret != global.ERROR_INSUFFICIENT_BUFFER {
			return result, syscall.Errno(ret)
		}
		buf = make([]byte, size)
	}
	ptr := (*MIB_UDPTABLE_OWNER_PID)(unsafe.Pointer(&buf[0]))
	for _, j := range ptr.AllUdp() {
		localPort, _, _ := procntohl.Call(uintptr(j.dwLocalPort))
		udpInfo := UPDInfo{
			PID:       j.dwOwningPid,
			LocalAddr: uint32ToIP(j.dwLocalAddr),
			LocalPort: uint16(localPort),
		}
		result[j.dwOwningPid] = append(result[j.dwOwningPid], udpInfo)
	}
	return result, nil
}

func DupicateHandle(pid uint32, handle syscall.Handle, hCurrent syscall.Handle) (syscall.Handle, error) {
	hProcess, _, _ := procOpenProcess.Call(
		global.PROCESS_DUP_HANDLE,
		0,
		uintptr(pid),
	)
	defer procCloseHandle.Call(hProcess)
	if hProcess == 0 {
		return 0, syscall.Errno(hProcess)
	}
	var dupHandle syscall.Handle
	ret, _, _ := procNtDuplicateObject.Call(
		hProcess,
		uintptr(handle),
		uintptr(hCurrent),
		uintptr(unsafe.Pointer(&dupHandle)),
		0,
		0,
		global.DUPLICATE_SAME_ACCESS,
	)
	if ret != 0 {
		return 0, syscall.Errno(ret)
	}
	return dupHandle, nil
}

func GetSocketAddresses(sock syscall.Handle) (localIP net.IP, localPort uint16, remoteIP net.IP, remotePort uint16, err error) {
	defer procCloseHandle.Call(uintptr(sock))

	var localRaw [128]byte
	localLen := int32(len(localRaw))

	ret, _, _ := procGetSockName.Call(
		uintptr(sock),
		uintptr(unsafe.Pointer(&localRaw[0])),
		uintptr(unsafe.Pointer(&localLen)),
	)

	if ret != 0 {
		err = errors.New("GetSockName error!")
		fmt.Println(err)
		return
	}

	saFamily := *(*uint16)(unsafe.Pointer(&localRaw[0]))

	switch saFamily {
	case syscall.AF_INET:
		sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(&localRaw[0]))
		ret, _, _ = procntohl.Call(uintptr(sa.Port))
		localPort = uint16(ret)

	case syscall.AF_INET6:
		sa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(&localRaw[0]))
		ret, _, _ = procntohl.Call(uintptr(sa.Port))
		remotePort = uint16(ret)
	}

	var remoteRaw [128]byte
	remoteLen := int32(len(remoteRaw))

	ret, _, _ = procGetpeername.Call(
		uintptr(sock),
		uintptr(unsafe.Pointer(&remoteRaw[0])),
		uintptr(unsafe.Pointer(&remoteLen)),
	)

	if ret != 0 {
		return
	}

	saFamily = *(*uint16)(unsafe.Pointer(&remoteRaw[0]))

	switch saFamily {
	case syscall.AF_INET:
		sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(&remoteRaw[0]))
		ret, _, _ = procntohl.Call(uintptr(sa.Port))
		localPort = uint16(ret)

	case syscall.AF_INET6:
		sa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(&remoteRaw[0]))
		ret, _, _ = procntohl.Call(uintptr(sa.Port))
		remotePort = uint16(ret)
	}

	return
}

func isSocket(handle syscall.Handle) bool {
	var size = uint32(5024)
	var buf = make([]byte, size)
	var sizeRt uint32
	for {
		ret, _, _ := procNtQueryObject.Call(
			uintptr(handle),
			global.ObjectNameInformation,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 0 {
			break
		}
		if ret != global.STATUS_INFO_LENGTH_MISMATCH {
			return false
		}
		size = sizeRt
		buf = make([]byte, size)
	}
	objectInfo := (*OBJECT_NAME_INFORMATION)(unsafe.Pointer(&buf[0]))
	if objectInfo.Name.Length == 0 {
		return false
	}
	name := UnicodeStringToString(objectInfo.Name)
	if name == `\Device\Afd` || name == `\Device\Udp` {
		return true
	}
	return false
}

func uint32ToIP(ip uint32) net.IP {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, ip)
	return net.IP(bytes)
}

func UnicodeStringToString(u UNICODE_STRING) string {
	length := u.Length / 2
	slice := (*[1 << 20]uint16)(unsafe.Pointer(u.Buffer))[:length:length]
	return syscall.UTF16ToString(slice[:])
}

//========================================================================

func EnableSeDebugPrivilege() error {
	var hToken syscall.Handle

	hProc, _, _ := procGetCurrentProcess.Call()
	ret, _, err := procOpenProcessToken.Call(
		hProc,
		uintptr(global.TOKEN_ADJUST_PRIVILEGES|global.TOKEN_QUERY),
		uintptr(unsafe.Pointer(&hToken)),
	)
	if ret == 0 {
		return fmt.Errorf("OpenProcessToken failed: %v", err)
	}
	defer syscall.CloseHandle(hToken)

	var luid LUID
	seDebugName, _ := syscall.UTF16PtrFromString("SeDebugPrivilege")
	ret, _, err = procLookupPrivilegeValueW.Call(
		0,
		uintptr(unsafe.Pointer(seDebugName)),
		uintptr(unsafe.Pointer(&luid)),
	)
	if ret == 0 {
		return fmt.Errorf("LookupPrivilegeValue failed: %v", err)
	}

	tp := TokenPrivileges{
		PrivilegeCount: 1,
		Privileges: [1]LUIDAndAttributes{{
			Luid:       luid,
			Attributes: global.SE_PRIVILEGE_ENABLED,
		}},
	}
	ret, _, _ = procAdjustTokenPrivileges.Call(
		uintptr(hToken),
		0,
		uintptr(unsafe.Pointer(&tp)),
		0,
		0,
		0,
	)

	lastErr := syscall.GetLastError()
	if lastErr != nil {
		if lastErr == syscall.Errno(global.ERROR_NOT_ALL_ASSIGNED) {
			return fmt.Errorf("SeDebugPrivilege not assigned (run as admin required)")
		}
		return fmt.Errorf("AdjustTokenPrivileges failed: %v", lastErr)
	}

	return nil
}
