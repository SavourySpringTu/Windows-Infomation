//go:build amd64
// +build amd64

package networkInfo

import (
	"unsafe"
)

func (g *MIB_IFTABLE) AllIfRow() []MIB_IFROW {
	return (*[(1 << 28) - 1]MIB_IFROW)(unsafe.Pointer(&g.table[0]))[:g.dwNumEntries:g.dwNumEntries]
}

func (g *MIB_IPNETTABLE) AllIPNetRow() []MIB_IPNETROW {
	return (*[(1 << 28) - 1]MIB_IPNETROW)(unsafe.Pointer(&g.table[0]))[:g.dwNumEntries:g.dwNumEntries]
}
