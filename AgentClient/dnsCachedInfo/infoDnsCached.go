package dnsCachedInfo

import (
	"golang.org/x/sys/windows"
	"net"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	DNS_QUERY_NO_WIRE_QUERY = 0x00000010
)

var (
	dnsapi                   = syscall.NewLazyDLL("Dnsapi.dll")
	procDnsGetCacheDataTable = dnsapi.NewProc("DnsGetCacheDataTable")
	procDnsQuery_W           = dnsapi.NewProc("DnsQuery_W")
)

type DNS_CACHE_ENTRY struct {
	pNext       uintptr
	pszName     *uint16
	wType       uint16
	wDataLength uint16
	dwFlags     uint32
}
type DNSRecord struct {
	pNext    uintptr
	pName    *uint16
	wType    uint16
	wDataLen uint16
	flags    uint32
	dwTtl    uint32
	reserved uint32
	Data     DNS_RECORD_DATA
}

type DNS_PTR_DATA struct {
	pNameHost *uint16
}

type DNS_RECORD_DATA struct {
	Data [16]byte
}

type RecordInfo struct {
	RecordName string
	Type       uint16
	Record     string
}

func GetDnsCachedInfo() (map[string][]RecordInfo, error) {
	var result = make(map[string][]RecordInfo)
	var pTable uintptr
	ret, _, err := procDnsGetCacheDataTable.Call(
		uintptr(unsafe.Pointer(&pTable)),
	)
	if ret == 0 {
		return result, err
	}

	for pTable != 0 {
		entry := (*DNS_CACHE_ENTRY)(unsafe.Pointer(pTable))
		domain := windows.UTF16PtrToString(entry.pszName)
		dnsCachedSlice, _ := GetDNSByDomain(domain, entry.wType)
		result[domain] = append(result[domain], dnsCachedSlice...)
		pTable = entry.pNext
	}
	return result, nil
}

func GetDNSByDomain(domain string, wType uint16) ([]RecordInfo, error) {
	var result []RecordInfo
	namePtr, _ := syscall.UTF16PtrFromString(domain)
	var pRecord uintptr

	ret, _, err := procDnsQuery_W.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(wType),
		uintptr(DNS_QUERY_NO_WIRE_QUERY),
		0,
		uintptr(unsafe.Pointer(&pRecord)),
		0,
	)
	if ret != 0 {
		return result, err
	}
	for pRecord != 0 {
		rec := (*DNSRecord)(unsafe.Pointer(pRecord))
		dnsCached := RecordInfo{
			RecordName: windows.UTF16PtrToString(rec.pName),
			Type:       rec.wType,
		}
		switch rec.wType {
		case 1:
			ip := strconv.Itoa(int(rec.Data.Data[0])) + "." + strconv.Itoa(int(rec.Data.Data[1])) + "." + strconv.Itoa(int(rec.Data.Data[2])) + "." + strconv.Itoa(int(rec.Data.Data[3]))
			dnsCached.Record = ip
		case 5:
			cNamePtr := (*DNS_PTR_DATA)(unsafe.Pointer(&rec.Data))
			cName := windows.UTF16PtrToString(cNamePtr.pNameHost)
			dnsCached.Record = cName
		case 28:
			ipSliceByte := rec.Data.Data
			ipV6Net := net.IP(ipSliceByte[:])
			dnsCached.Record = ipV6Net.String()
		}
		result = append(result, dnsCached)
		pRecord = rec.pNext
	}
	return result, nil
}
