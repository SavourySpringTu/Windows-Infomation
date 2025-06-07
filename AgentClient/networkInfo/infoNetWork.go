package networkInfo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"net"
	"syscall"
	"unsafe"
)

const (
	AF_UNSPEC                       = 0
	GAA_FLAG_INCLUDE_ALL_INTERFACES = 0x0100
)

const (
	ERROR_INSUFFICIENT_BUFFER = 122
)

var (
	iphlpapi                 = syscall.NewLazyDLL("Iphlpapi.dll")
	procGetAdaptersAddresses = iphlpapi.NewProc("GetAdaptersAddresses")
	procGetIfTable           = iphlpapi.NewProc("GetIfTable")
	procGetIpNetTable        = iphlpapi.NewProc("GetIpNetTable")
)

type MIB_IFROW struct {
	Name            [256]uint16
	Index           uint32
	Type            uint32
	Mtu             uint32
	Speed           uint32
	PhysAddrLen     uint32
	PhysAddr        [8]byte
	AdminStatus     uint32
	OperStatus      uint32
	LastChange      uint32
	InOctets        uint32
	InUcastPkts     uint32
	InNUcastPkts    uint32
	InDiscards      uint32
	InErrors        uint32
	InUnknownProtos uint32
	OutOctets       uint32
	OutUcastPkts    uint32
	OutNUcastPkts   uint32
	OutDiscards     uint32
	OutErrors       uint32
	OutQLen         uint32
	DescrLen        uint32
	Descr           [256]byte
}

type MIB_IFTABLE struct {
	dwNumEntries uint32
	table        [1]MIB_IFROW
}

type MIB_IPNETROW struct {
	dwIndex       uint32
	dwPhysAddrLen uint32
	bPhysAddr     [6]byte
	dwAddr        uint32
	dwType        uint32
}

type MIB_IPNETTABLE struct {
	dwNumEntries uint32
	table        [1]MIB_IPNETROW
}

type Octets struct {
	InOctets  uint32
	OutOctets uint32
}

type ARP struct {
	Index        uint32
	InterNetAddr net.IP
	PhysicalAddr string
	Type         string
}

type NetWorkInfo struct {
	Index        uint32
	Name         string
	MacAddress   string
	LocalAddress []net.IP
	Octets       Octets
	ARP          []ARP
}

var TypeARP = []string{
	"Other",
	"Invalid",
	"Dynamic",
	"Static",
}

// Get info netword interface
func GetInfoNetWork() ([]NetWorkInfo, error) {
	GetOctetsNetWorkInterface()
	var result []NetWorkInfo
	var size = uint32(15000)
	var buf = make([]byte, size)
	ret, _, err := procGetAdaptersAddresses.Call(
		AF_UNSPEC,
		0, 0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 {
		return result, err
	}
	ipAdap := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0]))
	buf = nil

	// Get octets
	mapOctets, errOctet := GetOctetsNetWorkInterface()
	if errOctet != nil {
		fmt.Println(errOctet)
	}

	// Get arp
	mapARP, errARP := GetARPTable()
	if errARP != nil {
		fmt.Println(errARP)
	}

	for ipAdap != nil {
		name := Uint16PtrToString(ipAdap.FriendlyName)

		//Convert [8]byte to [6]byte and then convert it to a MAC address.
		mac := SliceByteToMacAddr([6]byte(ipAdap.PhysicalAddress[:6]))
		var localAddSlice []net.IP

		// Get all address of network interface
		for ipAdap.FirstUnicastAddress != nil {
			localAddr, errParesAddr := ParseSockAddrAnyToNetIP(ipAdap.FirstUnicastAddress.Address.Sockaddr)
			if errParesAddr == nil {
				localAddSlice = append(localAddSlice, localAddr)
				ipAdap.FirstUnicastAddress = ipAdap.FirstUnicastAddress.Next
			}
		}
		netWord := NetWorkInfo{
			Index:        ipAdap.IfIndex,
			Name:         name,
			MacAddress:   mac,
			LocalAddress: localAddSlice,
			Octets:       mapOctets[ipAdap.IfIndex],
			ARP:          mapARP[ipAdap.IfIndex],
		}
		result = append(result, netWord)
		ipAdap = ipAdap.Next
	}
	return result, nil
}

// Get all Octets, return a map with the key as the index, value is an struct octet.
func GetOctetsNetWorkInterface() (map[uint32]Octets, error) {
	var result = make(map[uint32]Octets)
	var size = uint32(32)
	var buf = make([]byte, size)
	for {
		ret, _, _ := procGetIfTable.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
		)
		if ret == 0 {
			break
		}
		if ret != ERROR_INSUFFICIENT_BUFFER {
			return result, errors.New("GetIfTable Error!")
		}
		buf = make([]byte, size)
	}
	ptr := (*MIB_IFTABLE)(unsafe.Pointer(&buf[0]))
	allIfRow := ptr.AllIfRow()
	for _, i := range allIfRow {
		if i.InOctets != 0 || i.OutOctets != 0 {
			octets := Octets{
				InOctets:  i.InOctets,
				OutOctets: i.OutOctets,
			}
			result[i.Index] = octets
		}
	}
	return result, nil
}

// Get all ARP table
func GetARPTable() (map[uint32][]ARP, error) {
	var result = make(map[uint32][]ARP)
	var size = uint32(32)
	var buf = make([]byte, size)
	for {
		ret, _, _ := procGetIpNetTable.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
		)
		if ret == 0 {
			break
		}
		if ret != ERROR_INSUFFICIENT_BUFFER {
			return result, errors.New("GetIpNetTable error!")
		}
		buf = make([]byte, size)
	}

	ptr := (*MIB_IPNETTABLE)(unsafe.Pointer(&buf[0]))
	allIpNetRow := ptr.AllIPNetRow()
	for _, i := range allIpNetRow {
		arp := ARP{
			Index:        i.dwIndex,
			PhysicalAddr: SliceByteToMacAddr(i.bPhysAddr),
			InterNetAddr: Uint32ToIP(i.dwAddr),
			Type:         TypeARP[i.dwType-1], // type
		}
		result[i.dwIndex] = append(result[i.dwIndex], arp)
	}
	return result, nil
}

// Parse uint16 pointer to string
func Uint16PtrToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}
	var utf16Str []uint16
	for p := ptr; *p != 0; p = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + unsafe.Sizeof(*p))) {
		utf16Str = append(utf16Str, *p)
	}
	return syscall.UTF16ToString(utf16Str)
}

// Parse slice byte to string
func SliceByteToMacAddr(addr [6]byte) string {
	// The first 6 bytes
	result := fmt.Sprintf("%02X-%02X-%02X-%02X-%02X-%02X", addr[0], addr[1], addr[2], addr[3], addr[4], addr[5])
	return result
}

// Parse SockAddrAny to Net.IP
func ParseSockAddrAnyToNetIP(rsa *syscall.RawSockaddrAny) (net.IP, error) {
	switch rsa.Addr.Family {
	case syscall.AF_INET:
		sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
		return net.IPv4(sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3]), nil
	case syscall.AF_INET6:
		sa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(rsa))
		return net.IP(sa.Addr[:]), nil
	default:
		return nil, errors.New("Unknow!")
	}
}

// Parse uint32 to Net.IP
func Uint32ToIP(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, nn)
	return ip
}
