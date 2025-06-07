//go:build amd64
// +build amd64

package connectionsInfo

import (
	"unsafe"
)

func (g *MID_TCPTABLE_OWNER_PID) AllTcpTable() []MIB_TCPROW_OWNER_PID {
	return (*[(1 << 28) - 1]MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&g.TcpRowOwnerPID[0]))[:g.dwNumEntries:g.dwNumEntries]
}

func (g *MIB_UDPTABLE_OWNER_PID) AllUdpTable() []MIB_UDPROW_OWNER_PID {
	return (*[(1 << 28) - 1]MIB_UDPROW_OWNER_PID)(unsafe.Pointer(&g.UdpRowOwnerPID[0]))[:g.dwNumEntries:g.dwNumEntries]
}

func (g *SYSTEM_HANDLE_INFOMATION) AllHandle() []SYSTEM_HANDLE {
	return (*[(1 << 28) - 1]SYSTEM_HANDLE)(unsafe.Pointer(&g.Handles[0]))[:g.HandleCount:g.HandleCount]
}
