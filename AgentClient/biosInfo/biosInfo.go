package biosInfo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	firmwareTableProviderSigRSMB uint32 = 0x52534d42
	PhysicalDrive0                      = "\\\\.\\PhysicalDrive0"
)

const (
	GENERIC_READ                 = 0x80000000
	GENERIC_WRITE                = 0x40000000
	FILE_SHARE_READ              = 0x00000001
	FILE_SHARE_WRITE             = 0x00000002
	OPEN_EXISTING                = 3
	IOCTL_STORAGE_QUERY_PROPERTY = 0x002D1400
)

const (
	TableSystem    = 1
	TableBaseBoard = 2
	TableProcessor = 4

	SystemManufacturerFormat    = 0
	SystemModelFormat           = 1
	ProcessorNameFormat         = 12
	UUIDFormat                  = 4
	BaseBoardManufacturerFormat = 0
	BaseBoardModelFormat        = 1
	BaseBoardSerialFormat       = 3

	StorageDeviceProperty = 0
	PropertyStandardQuery = 0
)

var (
	kernel32                   = syscall.NewLazyDLL("Kernel32.dll")
	procGetSystemFirmwareTable = kernel32.NewProc("GetSystemFirmwareTable")
	procCreateFileW            = kernel32.NewProc("CreateFileW")
	procDeviceIoControl        = kernel32.NewProc("DeviceIoControl")
)

type RawSMBIOSData struct {
	Used20CallingMethod byte
	SMBIOSMajorVersion  byte
	SMBIOSMinorVersion  byte
	DmiRevision         byte
	Length              uint32
	SMBIOSTableData     []byte
}

type HeaderSMBIOS struct {
	Type   byte
	Length byte
	Handle uint16
}

type STORAGE_PROPERTY_QUERY struct {
	PropertyId uint32
	QueryType  uint32
	_          [2]byte
}

type STORAGE_DEVICE_DESCRIPTOR struct {
	Version               uint32
	Size                  uint32
	DeviceType            byte
	DeviceTypeModifier    byte
	RemovableMedia        bool
	CommandQueueing       bool
	VendorIdOffset        uint32
	ProductIdOffset       uint32
	ProductRevisionOffset uint32
	SerialNumberOffset    uint32
	BusType               uint32
	RawPropertiesLength   uint32
}

type BIOSInfo struct {
	SystemManufacturer    string
	SystemModel           string
	UUID                  string
	Processor             string
	BaseBoardManufacturer string
	BaseBoardSerial       string
	BaseBoardModel        string
	DiskDriveSerial       string
}

// Retrieve a buffer containing all table MSBIOS
func GetBiosInfo() (BIOSInfo, error) {
	var result BIOSInfo
	var size = uint32(32)
	var buf = make([]byte, size)
	for {
		ret, _, _ := procGetSystemFirmwareTable.Call(
			uintptr(firmwareTableProviderSigRSMB),
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
		)
		if uint32(ret) <= size {
			break
		}
		if ret == 0 {
			return result, errors.New("GetSystemFirmwareTable fail!")
		}
		size = uint32(ret)
		buf = make([]byte, size)
	}

	smDataTable := buf[8:size]

	// table system type 1
	tableSystem, errSystem := GetInfoTableSystem(smDataTable)
	if errSystem == nil {
		result.SystemManufacturer = tableSystem.SystemManufacturer
		result.SystemModel = tableSystem.SystemModel
		result.UUID = tableSystem.UUID
	}

	// table base board type 2
	tableBaseBoard, errBaseBoard := GetInfoTableBaseBoard(smDataTable)
	if errBaseBoard == nil {
		result.BaseBoardManufacturer = tableBaseBoard.BaseBoardManufacturer
		result.BaseBoardSerial = tableBaseBoard.BaseBoardSerial
		result.BaseBoardModel = tableBaseBoard.BaseBoardModel
	}

	// table processor type 4
	tableProcessor, errProcessor := GetInfoTableProcessor(smDataTable)
	if errProcessor == nil {
		result.Processor = tableProcessor.Processor
	}

	// Disk drive serial
	diskDriveSerial, errDiskDriveSerial := GetDiskDriveSerial()
	if errDiskDriveSerial == nil {
		result.DiskDriveSerial = diskDriveSerial
	}

	return result, nil
}

// Get info in table system
func GetInfoTableSystem(buff []byte) (BIOSInfo, error) {
	var result BIOSInfo

	table, err := GetTableByType(buff, TableSystem)
	if err != nil {
		return result, err
	}
	header := (*HeaderSMBIOS)(unsafe.Pointer(&table[0]))
	buffFormat := table[4:header.Length] // format area (skip 4byte of header)
	buffString := table[header.Length:]  // strings area

	// System Manufacturer
	systemManufacturer := getSMBiosStringByFormattedIndex(buffFormat, buffString, SystemManufacturerFormat)
	result.SystemManufacturer = systemManufacturer

	// System Model
	systemModel := getSMBiosStringByFormattedIndex(buffFormat, buffString, SystemModelFormat)
	result.SystemModel = systemModel

	// UUID
	buffUUID := buffFormat[UUIDFormat:20]
	uuid := parseUUID(buffUUID)
	result.UUID = uuid
	return result, nil
}

// Get infor in table base board
func GetInfoTableBaseBoard(buff []byte) (BIOSInfo, error) {
	var result BIOSInfo
	table, err := GetTableByType(buff, TableBaseBoard)
	if err != nil {
		return result, err
	}
	header := (*HeaderSMBIOS)(unsafe.Pointer(&table[0]))
	buffFormat := table[4:header.Length]
	buffString := table[header.Length:]

	// Base Board Manufacturer
	baseBoardManufacturerStr := getSMBiosStringByFormattedIndex(buffFormat, buffString, BaseBoardManufacturerFormat)
	result.BaseBoardManufacturer = baseBoardManufacturerStr

	// Base Board Serial
	baseBoardSerialStr := getSMBiosStringByFormattedIndex(buffFormat, buffString, BaseBoardSerialFormat)
	result.BaseBoardSerial = baseBoardSerialStr

	// Base Board Model
	baseBoardModelStr := getSMBiosStringByFormattedIndex(buffFormat, buffString, BaseBoardModelFormat)
	result.BaseBoardModel = baseBoardModelStr

	return result, nil
}

// Get info in table processor
func GetInfoTableProcessor(buff []byte) (BIOSInfo, error) {
	var result BIOSInfo
	table, err := GetTableByType(buff, TableProcessor)
	if err != nil {
		return result, err
	}
	header := (*HeaderSMBIOS)(unsafe.Pointer(&table[0]))
	buffFormat := table[4:header.Length]
	buffString := table[header.Length:]

	// get processor name
	processorStr := getSMBiosStringByFormattedIndex(buffFormat, buffString, ProcessorNameFormat)
	result.Processor = processorStr

	return result, nil
}

// Get buff table by type table
func GetTableByType(buf []byte, typeTable int) ([]byte, error) {
	start := 0
	if int(buf[0]) == typeTable {
		return buf, nil
	}
	header := (*HeaderSMBIOS)(unsafe.Pointer(&buf[0]))
	index := int(header.Length)
	for ; index < len(buf)-1; index++ {
		if buf[index] == 0x00 && buf[index+1] == 0x00 {
			header = (*HeaderSMBIOS)(unsafe.Pointer(&buf[start]))
			if int(header.Type) == typeTable {
				table := buf[start:] // Return the first element of the target table and continue until the end
				return table, nil
			}
			start = index + 2               // The first element of the next table
			index += 1 + int(header.Length) // Skip the header and format area.
		}
	}
	return nil, errors.New("Can't find")
}

// Get data in strings area by format index
func getSMBiosStringByFormattedIndex(buffFormat []byte, buffString []byte, offset int) string {
	index := buffFormat[offset]
	if index == 0 {
		return ""
	}
	currentIndex := byte(1)
	start := 0
	for i := 0; i < len(buffString); i++ {
		if buffString[i] == 0 {
			if currentIndex == index {
				return string(buffString[start:i])
			}
			currentIndex++
			start = i + 1

			if i+1 < len(buffString) && buffString[i+1] == 0 {
				break
			}
		}
	}
	return ""
}

// parse UUID form RFC 4122
// Data1 (0-3)
// Data2 (4-5)
// Data3 (6-7)
// Data4+5 (8-15)
func parseUUID(uuid []byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		binary.LittleEndian.Uint32(uuid[0:4]),
		binary.LittleEndian.Uint16(uuid[4:6]),
		binary.LittleEndian.Uint16(uuid[6:8]),
		uuid[8], uuid[9], uuid[10], uuid[11], uuid[12], uuid[13], uuid[14], uuid[15],
	)
}

// Get Disk Drive Serial Number
func GetDiskDriveSerial() (string, error) {
	driveUTF16Ptr, _ := syscall.UTF16PtrFromString(PhysicalDrive0)

	handle, _, errCreate := procCreateFileW.Call(
		uintptr(unsafe.Pointer(driveUTF16Ptr)),
		uintptr(GENERIC_READ|GENERIC_WRITE),
		uintptr(FILE_SHARE_READ),
		0,
		uintptr(OPEN_EXISTING),
		0, 0,
	)
	if handle == 0 {
		return "", errCreate
	}

	//Set up the request to send to the device.
	var query STORAGE_PROPERTY_QUERY
	query.PropertyId = StorageDeviceProperty
	query.QueryType = PropertyStandardQuery

	sizeBuf := uint32(50)
	buf := make([]byte, sizeBuf)
	var returned uint32
	for {
		ret, _, err := procDeviceIoControl.Call(
			handle,
			IOCTL_STORAGE_QUERY_PROPERTY,
			uintptr(unsafe.Pointer(&query)),
			unsafe.Sizeof(query),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(sizeBuf),
			uintptr(unsafe.Pointer(&returned)),
			0,
		)
		if sizeBuf > returned && ret == 1 {
			break
		} else if sizeBuf > returned && ret == 0 {
			return "", err
		}
		if sizeBuf == returned {
			sizeBuf = sizeBuf * 2
			buf = make([]byte, sizeBuf)
		}
	}
	descriptor := (*STORAGE_DEVICE_DESCRIPTOR)(unsafe.Pointer(&buf[0]))
	offset := descriptor.SerialNumberOffset
	if offset == 0 || offset > returned {
		return "", errors.New("Offset invalid!")
	}
	serial := ""
	// get data by offset
	for i := offset; i < sizeBuf && buf[i] != 0; i++ {
		serial += string(buf[i])
	}
	return serial, nil
}
