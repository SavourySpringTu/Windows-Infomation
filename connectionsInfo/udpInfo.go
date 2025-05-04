package connectionsInfo

import (
	"encoding/binary"
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

const ()

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

type UPDInfo struct {
	PID        uint32
	Handle     uint16
	RemoteAddr net.IP
	LocalAddr  net.IP
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
	var result = make(map[uint32][]UPDInfo)
	hCurrent, _, errCurrent := procGetCurrentProcess.Call()
	if hCurrent == 0 {
		return result, errCurrent
	}
	allHandle, errAllHandle := FilterHandle()
	if errAllHandle != nil {
		return result, errAllHandle
	}
	for _, handle := range allHandle {
		dupHandle, errDup := DupicateHandle(handle.PID, syscall.Handle(handle.Handle), syscall.Handle(hCurrent))
		if errDup != nil {
			continue
		}
		localAddr, remoteAddr, _ := GetSocketAddresses(dupHandle)
		udpInfo := UPDInfo{
			PID:        handle.PID,
			Handle:     handle.Handle,
			LocalAddr:  localAddr,
			RemoteAddr: remoteAddr,
		}
		result[handle.PID] = append(result[handle.PID], udpInfo)
	}
	procCloseHandle.Call(hCurrent)
	return result, nil
}

func FilterHandle() ([]UPDInfo, error) {
	var result []UPDInfo
	allHandle, errHandle := GetAllHandle()
	if errHandle != nil {
		return result, errHandle
	}
	allUdp, errUdp := GetAllUdp()
	if errUdp != nil {
		return result, errUdp
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
	return result, nil
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

func GetAllUdp() (map[uint32]struct{}, error) {
	var result = make(map[uint32]struct{})
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
		result[j.dwOwningPid] = struct{}{}
	}
	return result, nil
}

func DupicateHandle(pid uint32, handle syscall.Handle, hCurrent syscall.Handle) (syscall.Handle, error) {
	hProcess, _, _ := procOpenProcess.Call(
		global.PROCESS_DUP_HANDLE,
		1,
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
		global.DUPLICATE_SAME_ACCESS,
	)
	if ret == 0 {
		return 0, syscall.Errno(ret)
	}
	return dupHandle, nil
}

func GetSocketAddresses(sock syscall.Handle) (localAddr, remoteAddr net.IP, err error) {
	defer procCloseHandle.Call(uintptr(sock))

	var remote Sockaddr
	var remoteSize = int32(unsafe.Sizeof(remote))

	ret, _, _ := procGetpeername.Call(
		uintptr(sock),
		uintptr(unsafe.Pointer(&remote)),
		uintptr(unsafe.Pointer(&remoteSize)),
	)
	if ret != 0 {
		errCode, _, _ := procWSAGetLastErr.Call()
		return nil, nil, syscall.Errno(errCode)
	}

	var local Sockaddr
	var localSize = int32(unsafe.Sizeof(local))

	ret, _, _ = procGetSockName.Call(
		uintptr(sock),
		uintptr(unsafe.Pointer(&local)),
		uintptr(unsafe.Pointer(&localSize)),
	)
	if ret != 0 {
		errCode, _, _ := procWSAGetLastErr.Call()
		return nil, nil, syscall.Errno(errCode)
	}

	remoteAddr = net.IPv4(
		remote.Address.Addr[0],
		remote.Address.Addr[1],
		remote.Address.Addr[2],
		remote.Address.Addr[3],
	)

	localAddr = net.IPv4(
		local.Address.Addr[0],
		local.Address.Addr[1],
		local.Address.Addr[2],
		local.Address.Addr[3],
	)

	return localAddr, remoteAddr, nil
}

func uint32ToIP(ip uint32) net.IP {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, ip)
	return net.IP(bytes)
}
