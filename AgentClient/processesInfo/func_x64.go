//go:build amd64
// +build amd64

package processesInfo

import (
	"unsafe"
)

func (g *TOKEN_GROUPS) AllGroups() []SID_AND_ATTRIBUTES {
	return (*[(1 << 28) - 1]SID_AND_ATTRIBUTES)(unsafe.Pointer(&g.Groups[0]))[:g.GroupCount:g.GroupCount]
}

func (g *TOKEN_PRIVILEGES) AllPrivileges() []LUID_AND_ATTRIBUTES {
	return (*[(1 << 28) - 1]LUID_AND_ATTRIBUTES)(unsafe.Pointer(&g.Privileges[0]))[:g.PrivilegeCount:g.PrivilegeCount]
}
