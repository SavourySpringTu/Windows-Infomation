//go:build 386
// +build 386

package networkInfo

import (
	"unsafe"
)

func (g *MIB_IFTABLE) AllIfRow() []MIB_IFROW {
	return (*[(1 << 20) - 1]MIB_IFROW)(unsafe.Pointer(&g.table[0]))[:g.dwNumEntries:g.dwNumEntries]
}

func (g *MIB_IPNETTABLE) AllIPNetRow() []MIB_IPNETROW {
	return (*[(1 << 20) - 1]MIB_IPNETROW)(unsafe.Pointer(&g.table[0]))[:g.dwNumEntries:g.dwNumEntries]
}
