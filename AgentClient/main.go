package main

import (
	"AgentClient/appsInfo"
	"AgentClient/biosInfo"
	"AgentClient/connectionsInfo"
	"AgentClient/dnsCachedInfo"
	"AgentClient/filesInfo"
	"AgentClient/hardWareInfo"
	"AgentClient/hashFile"
	"AgentClient/kernelModuleInfo"
	"AgentClient/networkInfo"
	"AgentClient/processesInfo"
	"AgentClient/registryInfo"
	"AgentClient/sysInfo"
	"flag"
	"fmt"
)

var (
	processInfoFlag  = flag.String("p", "", "Get info of process")
	hardwareInfoFlag = flag.Bool("i", false, "Get info of hardware")
	registryInfoFlag = flag.String("r", "", "Get info of registry by path")
	fileInfoFlag     = flag.String("f", "", "Get info of file by path")
	fileHashFlag     = flag.String("h", "", "Get hash of file by path")
	sysInfoFlag      = flag.Bool("s", false, "Get name, version, build of OS")
	appInfoFlag      = flag.Bool("a", false, "Get Info of App")
	kernelModuleFlag = flag.Bool("k", false, "Get kernel module information")
	dnsCachedFlag    = flag.Bool("d", false, "Get dns cached")
	connectionFlag   = flag.Bool("c", false, "Get connections")
	biosInfoFlag     = flag.Bool("b", false, "Get bios information")
	netWorkInfoFlag  = flag.Bool("n", false, "Get network information")
)

func main() {
	flag.Parse()
	if *sysInfoFlag {
		userName, errUserName := sysInfo.GetUserName()
		if errUserName == nil {
			fmt.Printf("User name: %s\n", userName)
		} else {
			fmt.Println(errUserName)
		}
		namePC, errNamePC := sysInfo.GetNamePC()
		if errNamePC == nil {
			fmt.Printf("Name PC: %s\n", namePC)
		} else {
			fmt.Println(errNamePC)
		}
		nameOS2, errNameOS2 := sysInfo.GetInfoOSbyName("ProductName")
		if errNameOS2 == nil {
			fmt.Printf("Name OS: %s \n", nameOS2)
		} else {
			fmt.Println(errNameOS2)
		}

		versionOS, errVersionOS := sysInfo.GetInfoOSbyName("DisplayVersion")
		if errVersionOS == nil {
			fmt.Printf("Version OS: %s \n", versionOS)
		} else {
			fmt.Println(errVersionOS)
		}

		buildOS, errBuildOS := sysInfo.GetInfoOSbyName("CurrentBuild")
		if errBuildOS == nil {
			fmt.Printf("Build OS: %s \n", buildOS)
		} else {
			fmt.Println(errBuildOS)
		}
		archOS, errArchOS := sysInfo.GetArchitectureOS()
		if errArchOS == nil {
			fmt.Printf("Architecture OS: %s\n", archOS)
		} else {
			fmt.Println(errArchOS)
		}
		timeUp, errTimeUp := sysInfo.GetTimeUp()
		if errTimeUp == nil {
			fmt.Printf("Time up: %d minute\n", timeUp)
		} else {
			fmt.Println(errTimeUp)
		}

	}
	if *appInfoFlag {
		sliceAppInfo, errInfoApp := appsInfo.GetInfoApp()

		fmt.Println("======================== IN ============================")
		if errInfoApp == nil {
			for _, item := range sliceAppInfo {
				fmt.Printf("========== %s ==========\n", item.Name)
				fmt.Printf("Version: %s\n", item.Version)
				fmt.Printf("Publisher: %s\n", item.Publisher)
				fmt.Printf("Install Date: %s\n", item.InstallDate)
			}
		} else {
			fmt.Println("Error:", errInfoApp)
		}
	}
	//if *processInfoFlag == "" {
	//	sliceProcessInfo, err := processesInfo.GetInfoProcesses(0, true)
	//	if err != nil {
	//		fmt.Println(err)
	//	} else {
	//		PrintProcessesInfo(sliceProcessInfo)
	//	}
	//}
	//else {
	//	var pid uint64
	//	pid, _ = strconv.ParseUint(*processInfoFlag, 10, 32)
	//
	//	sliceProcessInfo, err := processesInfo.GetInfoProcesses(uint32(pid), false)
	//	if err != nil {
	//		fmt.Println(err)
	//	} else {
	//		PrintProcessesInfo(sliceProcessInfo)
	//	}
	//}

	if *hardwareInfoFlag {
		cpu, errCPU := hardWareInfo.GetNameCPU()
		if errCPU != nil {
			fmt.Println(errCPU)
		} else {
			fmt.Printf("CPU: %s\n", cpu)
		}
		ram, errRam := hardWareInfo.GetInfoRAM()
		if errRam != nil {
			fmt.Println(errRam)
		} else {
			fmt.Printf("RAM: %d Mb\n", ram)
		}
		disk, errDisk := hardWareInfo.GetSizeDisk()
		if errDisk != nil {
			fmt.Println(errDisk)
		} else {
			fmt.Printf("Disk: %d Mb\n", disk)
		}
	}
	if *registryInfoFlag != "" {
		mapRegistry, errMapRegistry := registryInfo.GetInfoRegistryByPath(*registryInfoFlag)
		if errMapRegistry != nil {
			fmt.Println(errMapRegistry)
		} else {
			for key, value := range mapRegistry {
				fmt.Println("=============================")
				fmt.Println("Name  : ", key)
				fmt.Println("Value : ", value)
			}
		}
	}
	if *fileInfoFlag != "" {
		fileInfo, err := filesInfo.GetInfoFileAndFolder(*fileInfoFlag)
		if err != nil {
			fmt.Println(err)
		} else {
			for _, i := range fileInfo {
				fmt.Println("====================================================")
				fmt.Println("Name         : ", i.Name)
				fmt.Println("Date create  : ", i.DateCreated)
				fmt.Println("Date modified: ", i.DateModified)
				fmt.Println("Size         : ", i.Size)
				fmt.Println("MD5          : ", i.MD5)
				fmt.Println("SHA1         : ", i.SHA1)
				fmt.Println("SHA256       : ", i.SHA_256)
			}
		}
	}
	if *fileHashFlag != "" {
		fileHash, err := hashFile.GetHashFileAndFolder(*fileHashFlag)
		if err != nil {
			fmt.Println(err)
		} else {
			for _, i := range fileHash {
				fmt.Println("====================================================")
				fmt.Println("Name    : ", i.Name)
				fmt.Println("MD5     : ", i.MD5)
				fmt.Println("SHA1    : ", i.SHA1)
				fmt.Println("SHA_256 : ", i.SHA_256)
			}
		}
	}
	if *kernelModuleFlag {
		kernelModule, err := kernelModuleInfo.GetInfoKernelModule()
		if err != nil {
			fmt.Println(err)
		} else {
			for _, i := range kernelModule {
				fmt.Println("====================================================")
				fmt.Println("Name        : ", i.Name)
				fmt.Println("Path        : ", i.Path)
				fmt.Println("Startup Mode: ", i.StartupMode)
				fmt.Println("State       : ", i.State)
				fmt.Println("SHA256      : ", i.SHA256)
			}
		}
	}
	if *dnsCachedFlag {
		dnsCachedInfoSlice, errDns := dnsCachedInfo.GetDnsCachedInfo()
		if errDns != nil {
			fmt.Println(errDns)
		} else {
			for i, j := range dnsCachedInfoSlice {
				fmt.Println("Domain: ", i)
				for _, k := range j {
					fmt.Println("	-----------------------------------")
					fmt.Println("	Record Name: ", k.RecordName)
					fmt.Println("	Type       : ", k.Type)
					fmt.Println("	Record     : ", k.Record)

				}
				fmt.Println("======================================================")
			}
		}
	}
	if *connectionFlag {
		connectionsTcpInfo, errTcp := connectionsInfo.GetTcpInfo()
		if errTcp == nil {
			fmt.Println("======================= TCP =========================")
			for i, j := range connectionsTcpInfo {
				fmt.Println("PID           : ", i)
				for _, k := range j {
					fmt.Println("Local Address : ", k.LocalAddress, ":", k.LocalPort)
					fmt.Println("Remote Address: ", k.RemoteAddress, ":", k.RemotePort)
					fmt.Println("State         : ", k.State)
					fmt.Println("---------------------------------------------------")
				}
			}
		}
		connectionUdpInfo, errUdp := connectionsInfo.GetUdpInfo()
		if errUdp == nil {
			fmt.Println("======================= UDP =========================")
			for i, j := range connectionUdpInfo {
				fmt.Println("PID  : ", i)
				for _, k := range j {
					fmt.Println("	Local Address : ", k.LocalAddr, ":", k.LocalPort)
					fmt.Println("	Remote Address: ", k.RemoteAddr, ":", k.RemotePort)
					fmt.Println("	---------------------------------------------------")
				}
				fmt.Println("==========================================")
			}
		}
	}
	if *biosInfoFlag {
		biosInfo, errBios := biosInfo.GetBiosInfo()
		if errBios == nil {
			fmt.Println("System Manufacturer    : ", biosInfo.SystemManufacturer)
			fmt.Println("System Model           : ", biosInfo.SystemModel)
			fmt.Println("UUID                   : ", biosInfo.UUID)
			fmt.Println("Processor              : ", biosInfo.Processor)
			fmt.Println("Base Board Manufacturer: ", biosInfo.BaseBoardManufacturer)
			fmt.Println("Base Board Serial      : ", biosInfo.BaseBoardSerial)
			fmt.Println("Base Board Model       : ", biosInfo.BaseBoardModel)
			fmt.Println("Disk Drive Serial      : ", biosInfo.DiskDriveSerial)
		}
	}
	if *netWorkInfoFlag {
		netWork, errNetWork := networkInfo.GetInfoNetWork()
		if errNetWork == nil {
			for _, i := range netWork {
				fmt.Println("Name           :", i.Name)
				fmt.Println("Mac Address    :", i.MacAddress)
				fmt.Println("Octets received:", i.Octets.InOctets)
				fmt.Println("Octets sent    :", i.Octets.OutOctets)
				fmt.Println("Local Address  ")
				for _, j := range i.LocalAddress {
					fmt.Println("  ", j)
				}
				fmt.Println("ARP")
				for _, j := range i.ARP {
					fmt.Println("----------------------------------------------")
					fmt.Println("  Internet Address: ", j.InterNetAddr)
					fmt.Println("  Physical Address: ", j.PhysicalAddr)
					fmt.Println("  Type            : ", j.Type)
				}
				fmt.Println("=================================")
			}
		}
	}
}

func PrintProcessesInfo(sliceProcessInfo []processesInfo.ProcessInfo) {
	for _, i := range sliceProcessInfo {
		fmt.Println("Pid           : ", i.Pid)
		fmt.Println("Name          : ", i.Name)
		fmt.Println("User          : ", i.Token.User)
		fmt.Println("Pid Parent    : ", i.ParentPid)
		fmt.Println("Path          : ", i.Path)
		fmt.Println("Commandline   : ", i.Commandline)
		fmt.Println("Runtime       : ", i.Runtime, " Millisecond")
		fmt.Println("SID           : ", i.Token.SID)
		fmt.Println("Session       : ", i.Token.Session)
		fmt.Println("Logon Session : ", i.Token.LogonSession)
		fmt.Println("Virtualized   : ", i.Token.Virtualized)
		fmt.Println("Protected     : ", i.Token.Protected)
		for _, j := range i.ConnectionTcp {
			fmt.Println("Connection TCP:")
			fmt.Println("	Local Address : ", j.LocalAddress)
			fmt.Println("	Remote Address: ", j.RemoteAddress)
			fmt.Println("	State         : ", j.State)
		}
		for _, j := range i.ConnectionUdp {
			fmt.Println("Connection UDP:")
			fmt.Println("	Handle        : ", j.Handle)
			fmt.Println("	Local Address : ", j.LocalAddr)
			fmt.Println("	Remote Address: ", j.RemoteAddr)
		}
		fmt.Println("Groups        :")
		for _, j := range i.Token.Groups {
			fmt.Println("	", j.Name, " : ", j.SID)
		}
		fmt.Println("Privileges    :")
		for _, j := range i.Token.Privileges {
			fmt.Println("	", j.LUID, ": ", j.Name)
		}
		fmt.Println("Module        :")
		for _, j := range i.Module {
			fmt.Println("	", j.Name)
		}
		fmt.Println("====================================================================================================")
	}
}
