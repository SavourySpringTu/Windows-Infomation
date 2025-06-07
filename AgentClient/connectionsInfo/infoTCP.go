package connectionsInfo

import (
	"encoding/binary"
	"net"
	"syscall"
	"unsafe"
)

var (
	iphlpapi                = syscall.NewLazyDLL("Iphlpapi.dll")
	procGetExtendedTcpTable = iphlpapi.NewProc("GetExtendedTcpTable")
	procntohs               = ws2_32.NewProc("ntohs")
)

const (
	AF_INET                         = 2
	TCP_TABLE_OWNER_PID_CONNECTIONS = 4
)

type MID_TCPTABLE_OWNER_PID struct {
	dwNumEntries   uint32
	TcpRowOwnerPID [1]MIB_TCPROW_OWNER_PID
}

type MIB_TCPROW_OWNER_PID struct {
	dwState      uint32
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

type ConnectionTcpInfo struct {
	Pid           uint32
	LocalAddress  net.IP
	RemoteAddress net.IP
	LocalPort     uint16
	RemotePort    uint16
	State         string
}

type Connections struct {
	ConnectionTcpInfo map[uint32][]ConnectionTcpInfo
	ConnectionUdpInfo map[uint32][]ConnectionUdpInfo
}

var State = []string{
	"MIB_TCP_STATE_CLOSED",
	"MIB_TCP_STATE_LISTEN",
	"MIB_TCP_STATE_SYN_SENT",
	"MIB_TCP_STATE_SYN_RCVD",
	"MIB_TCP_STATE_ESTAB",
	"MIB_TCP_STATE_FIN_WAIT1",
	"MIB_TCP_STATE_FIN_WAIT2",
	"MIB_TCP_STATE_CLOSE_WAIT",
	"MIB_TCP_STATE_CLOSING",
	"MIB_TCP_STATE_LAST_ACK",
	"MIB_TCP_STATE_TIME_WAIT",
	"MIB_TCP_STATE_DELETE_TCB",
}

func GetConnectionsInfo() (Connections, error) {
	tcp, _ := GetTcpInfo()
	udp, _ := GetUdpInfo()
	conn := Connections{
		ConnectionTcpInfo: tcp,
		ConnectionUdpInfo: udp,
	}
	return conn, nil
}

func GetTcpInfo() (map[uint32][]ConnectionTcpInfo, error) {
	var result = make(map[uint32][]ConnectionTcpInfo)
	var size = uint32(64)
	var buf = make([]byte, size)
	for {
		ret, _, _ := procGetExtendedTcpTable.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
			uintptr(AF_INET),
			TCP_TABLE_OWNER_PID_CONNECTIONS,
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
	tcpTable := (*MID_TCPTABLE_OWNER_PID)(unsafe.Pointer(&buf[0]))
	for _, i := range tcpTable.AllTcpTable() {
		ipLocal := Uint32ToIP(i.dwLocalAddr)
		ipRemote := Uint32ToIP(i.dwRemoteAddr)
		localPort, _, _ := procntohs.Call(uintptr(i.dwLocalPort))
		remotePort, _, _ := procntohs.Call(uintptr(i.dwRemotePort))
		connectionTcpInfo := ConnectionTcpInfo{
			Pid:           i.dwOwningPid,
			LocalAddress:  ipLocal,
			RemoteAddress: ipRemote,
			LocalPort:     uint16(localPort),
			RemotePort:    uint16(remotePort),
			State:         State[int(i.dwState)-1],
		}
		result[i.dwOwningPid] = append(result[i.dwOwningPid], connectionTcpInfo)
	}
	return result, nil
}

func Uint32ToIP(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, nn)
	return ip
}
