package connectionsInfo

import (
	"errors"
	"net"
	"syscall"
	"unsafe"
)

var (
	kernel32                     = syscall.NewLazyDLL("Kernel32.dll")
	ntdll                        = syscall.NewLazyDLL("ntdll.dll")
	ws2_32                       = syscall.NewLazyDLL("Ws2_32.dll")
	procGetExtendedUdpTable      = iphlpapi.NewProc("GetExtendedUdpTable")
	procNtQuerySystemInformation = ntdll.NewProc("NtQuerySystemInformation")
	procNtQueryObject            = ntdll.NewProc("NtQueryObject")
	procNtDuplicateObject        = ntdll.NewProc("NtDuplicateObject")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procGetCurrentProcess        = kernel32.NewProc("GetCurrentProcess")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procGetpeername              = ws2_32.NewProc("getpeername")
	procGetsockname              = ws2_32.NewProc("getsockname")
	procWSAEnumNetworkEvents     = ws2_32.NewProc("WSAEnumNetworkEvents")
)

const (
	ERROR_INSUFFICIENT_BUFFER   = 122
	SystemHandleInformation     = 16
	ObjectNameInformation       = 1
	STATUS_INFO_LENGTH_MISMATCH = 0xC0000004
	PROCESS_DUP_HANDLE          = 0x0040
	DUPLICATE_SAME_ACCESS       = 0x00000002
	UDP_TABLE_OWNER_PID         = 1
	ObjectTypeInformation       = 2
)

type SYSTEM_HANDLE struct {
	PID              uint32
	ObjectTypeNumber byte
	Flags            byte
	Handle           uint16
	Object           uintptr
	GrantedAccess    uint32
}

type SYSTEM_HANDLE_INFOMATION struct {
	HandleCount uint32
	Handles     [1]SYSTEM_HANDLE
}

type OBJECT_NAME_INFORMATION struct {
	Name UNICODE_STRING
}

type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

type MIB_UDPTABLE_OWNER_PID struct {
	dwNumEntries   uint32
	UdpRowOwnerPID [1]MIB_UDPROW_OWNER_PID
}

type MIB_UDPROW_OWNER_PID struct {
	dwLocalAddr uint32
	dwLocalPort uint32
	dwOwningPid uint32
}

type sockaddr_in struct {
	Family uint16
	Port   uint16
	Addr   [4]byte
	Zero   [8]byte
}

type ConnectionUdpInfo struct {
	Pid        uint32
	Handle     uint16
	LocalAddr  net.IP
	RemoteAddr net.IP
	LocalPort  uint16
	RemotePort uint16
}

func GetUdpInfo() (map[uint32][]ConnectionUdpInfo, error) {
	filterHandle, result, _ := FilterHandle()
	hCurrent, _, err := procGetCurrentProcess.Call()
	if hCurrent == 0 {
		return result, err
	}
	for _, i := range filterHandle {
		dupHandle, errDup := duplicateHandle(i.PID, syscall.Handle(i.Handle), syscall.Handle(hCurrent))
		if errDup != nil || isSocket(dupHandle) == false {
			procCloseHandle.Call(uintptr(dupHandle))
			continue
		}
		localAddr, localPort, remoteAddr, remotePort, _ := getSocketAddresses(dupHandle)
		connectionInfo := ConnectionUdpInfo{
			Pid:        i.PID,
			LocalAddr:  localAddr,
			RemoteAddr: remoteAddr,
			LocalPort:  localPort,
			RemotePort: remotePort,
		}
		result[i.PID] = append(result[i.PID], connectionInfo)

	}
	return result, nil
}

func FilterHandle() ([]SYSTEM_HANDLE, map[uint32][]ConnectionUdpInfo, error) {
	var result []SYSTEM_HANDLE
	allHandle, _ := GetAllHandle()
	mapCheck, _ := GetAllUdp()
	for _, i := range allHandle {
		if _, exists := mapCheck[i.PID]; exists {
			result = append(result, i)
		}
	}
	return result, mapCheck, nil
}

func GetAllUdp() (map[uint32][]ConnectionUdpInfo, error) {
	var result = make(map[uint32][]ConnectionUdpInfo)
	var size = uint32(64)
	var buf = make([]byte, size)
	for {
		ret, _, _ := procGetExtendedUdpTable.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
			uintptr(AF_INET),
			UDP_TABLE_OWNER_PID,
			0,
		)
		if ret == 0 {
			break
		}
		if ret == 122 {
			buf = make([]byte, size)
		} else if ret != 0 {
			return result, syscall.Errno(ret)
		}
	}
	ptr := (*MIB_UDPTABLE_OWNER_PID)(unsafe.Pointer(&buf[0]))
	for _, i := range ptr.AllUdpTable() {
		localPort, _, _ := procntohs.Call(uintptr(i.dwLocalPort))
		connectionInfo := ConnectionUdpInfo{
			Pid:       i.dwOwningPid,
			LocalAddr: Uint32ToIP(i.dwLocalAddr),
			LocalPort: uint16(localPort),
		}
		result[i.dwOwningPid] = append(result[i.dwOwningPid], connectionInfo)
	}
	return result, nil
}

func GetAllHandle() ([]SYSTEM_HANDLE, error) {
	var result []SYSTEM_HANDLE
	var size = uint32(64)
	var buf = make([]byte, size)
	var sizeRt uint32
	for {
		ret, _, _ := procNtQuerySystemInformation.Call(
			SystemHandleInformation,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 0 {
			break
		}
		if ret != STATUS_INFO_LENGTH_MISMATCH {
			return result, syscall.Errno(ret)
		}
		size = sizeRt
		buf = make([]byte, size)
	}
	ptr := (*SYSTEM_HANDLE_INFOMATION)(unsafe.Pointer(&buf[0]))
	result = ptr.AllHandle()
	return result, nil
}

func duplicateHandle(pid uint32, handle syscall.Handle, hCurrent syscall.Handle) (syscall.Handle, error) {
	hProcess, _, errProcess := procOpenProcess.Call(
		PROCESS_DUP_HANDLE,
		0,
		uintptr(pid),
	)

	defer procCloseHandle.Call(hProcess)
	if hProcess == 0 {
		return 0, errProcess
	}

	var dupHandle syscall.Handle
	ret, _, _ := procNtDuplicateObject.Call(
		hProcess,
		uintptr(handle),
		uintptr(hCurrent),
		uintptr(unsafe.Pointer(&dupHandle)),
		0,
		0,
		DUPLICATE_SAME_ACCESS,
	)
	if ret != 0 {
		return 0, syscall.Errno(ret)
	}
	return dupHandle, nil
}

func getSocketAddresses(sock syscall.Handle) (localIP net.IP, localPort uint16, remoteIP net.IP, remotePort uint16, err error) {
	defer procCloseHandle.Call(uintptr(sock))

	var localRaw syscall.RawSockaddrAny
	var remoteRaw syscall.RawSockaddrAny
	localLen := int32(unsafe.Sizeof(localRaw))
	remoteLen := int32(unsafe.Sizeof(remoteRaw))

	ret, _, callErr := procGetsockname.Call(
		uintptr(sock),
		uintptr(unsafe.Pointer(&localRaw)),
		uintptr(unsafe.Pointer(&localLen)),
	)
	if ret != 0 {
		err = callErr
		return
	}

	localIP, localPort, err = parseSockAddrAndPort(&localRaw)
	if err != nil {
		return
	}

	ret, _, callErr = procGetpeername.Call(
		uintptr(sock),
		uintptr(unsafe.Pointer(&remoteRaw)),
		uintptr(unsafe.Pointer(&remoteLen)),
	)
	if ret != 0 {
		if errno, ok := callErr.(syscall.Errno); ok {
			if errno == 10057 || errno == 10022 {
				remoteIP = nil
				remotePort = 0
				err = nil
				return
			}
		}
		err = callErr
		return
	}
	remoteIP, remotePort, err = parseSockAddrAndPort(&remoteRaw)
	return
}

func parseSockAddrAndPort(rsa *syscall.RawSockaddrAny) (net.IP, uint16, error) {
	switch rsa.Addr.Family {
	case syscall.AF_INET:
		sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
		port, _, _ := procntohs.Call(uintptr(sa.Port))
		return net.IPv4(sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3]), uint16(port), nil
	case syscall.AF_INET6:
		sa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(rsa))
		port, _, _ := procntohs.Call(uintptr(sa.Port))
		return net.IP(sa.Addr[:]), uint16(port), nil
	default:
		return nil, 0, errors.New("Unknow!")
	}
}

func isSocket(handle syscall.Handle) bool {
	var size = uint32(1)
	var buf = make([]byte, size)
	var sizeRt uint32
	for {
		ret, _, _ := procNtQueryObject.Call(
			uintptr(handle),
			uintptr(1),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&sizeRt)),
		)
		if ret == 0 {
			break
		}
		if ret != STATUS_INFO_LENGTH_MISMATCH {
			return false
		}
		size = sizeRt
		buf = make([]byte, size)
	}
	objectType := (*OBJECT_NAME_INFORMATION)(unsafe.Pointer(&buf[0]))
	if objectType.Name.Length == 0 {
		return false
	}
	dd := (*UNICODE_STRING)(unsafe.Pointer(&objectType.Name))
	slice := unsafe.Slice(dd.Buffer, dd.Length/2)
	nameStr := syscall.UTF16ToString(slice[:])
	if nameStr == `\Device\Afd` {
		return true
	}
	return false
}
