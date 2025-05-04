package main

import (
	"flag"
	"fmt"
	"main/appsInfo"
	"main/connectionsInfo"
	"main/filesInfo"
	"main/kernelModuleInfo"
	"main/processesInfo"
	"main/sysInfo"
)

var (
	sysInfoFlag       = flag.Bool("s", false, "Get information of system")
	appsInfoFlag      = flag.Bool("a", false, "Get information of all applications installed in the os")
	processesInfoFlag = flag.Bool("p", false, "Get infomation of processes")
	fileInfoFlag      = flag.Bool("f", false, "Get information of files")
	kernelModuleFlag  = flag.Bool("k", false, "Get information of kernel modules")
	connectionsFlag   = flag.Bool("c", false, "Get information of connections")
)

func main() {
	flag.Parse()
	if *sysInfoFlag {
		username, _ := sysInfo.GetUsername()
		fmt.Printf("User name: %s\n", username)

		nameOS, _ := sysInfo.GetInfoOSbyName("ProductName")
		fmt.Printf("Name OS: %s\n", nameOS)

		versionOS, _ := sysInfo.GetInfoOSbyName("DisplayVersion")
		fmt.Printf("Version OS: %s\n", versionOS)

		buildOS, _ := sysInfo.GetInfoOSbyName("CurrentBuild")
		fmt.Printf("Build OS: %s\n", buildOS)

		timeup, _ := sysInfo.GetUptime()
		fmt.Printf("Time up: %d minute\n", timeup)
	}
	if *appsInfoFlag {
		appInfo, _ := appsInfo.GetAllAppInfo()
		for _, item := range appInfo {
			fmt.Println("============================================================")
			fmt.Println("Name        : ", item.Name)
			fmt.Println("Version     : ", item.Version)
			fmt.Println("Publisher   : ", item.Publisher)
			fmt.Println("Install Date: ", item.InstallDate)
		}
	}

	if *processesInfoFlag {
		processInfo, errProcessInfo := processesInfo.GetProcessesInfo()
		if errProcessInfo != nil {
			fmt.Println(errProcessInfo)
		} else {
			for _, item := range processInfo {
				fmt.Println("==========================================================")
				fmt.Println("Pid          : ", item.Pid)
				fmt.Println("Name         : ", item.Name)
				fmt.Println("Pid parent   : ", item.PidParent)
				fmt.Println("SID          : ", item.Token.SID)
				fmt.Println("Session Id   : ", item.Token.SessionId)
				fmt.Println("Logon Session: ", item.Token.LogonSession)
				fmt.Println("Groups: ")
				for _, j := range item.Token.Groups {
					fmt.Println("	Sid: ", j.SID)
					fmt.Println("	------------------------------")
				}
				fmt.Println("Privileges: ")
				for _, j := range item.Token.Privileges {
					fmt.Println("	Name Privilege: ", j.Name)
					fmt.Println("	------------------------------")
				}
			}
		}
	}
	if *fileInfoFlag {
		filesInfo, errFilesInfo := filesInfo.GetInfoFilesAndFolder(`C:\Tu\Windows-Infomation\`)
		if errFilesInfo != nil {
			fmt.Println(errFilesInfo)
		} else {
			for _, item := range filesInfo {
				fmt.Println("=====================================")
				fmt.Println("Name          : ", item.Name)
				fmt.Println("Date created  : ", item.DateCreated)
				fmt.Println("Date modified : ", item.DateModified)
				fmt.Println("Size          : ", item.Size, " bytes")
			}
		}
	}
	if *kernelModuleFlag {
		kernelModuleInfo, _ := kernelModuleInfo.GetKernelModuleInfo()
		for _, item := range kernelModuleInfo {
			fmt.Println("==========================================================")
			fmt.Println("Name         : ", item.Name)
			fmt.Println("Path         : ", item.Path)
			fmt.Println("Status       : ", item.Status)
		}
	}
	if *connectionsFlag {
		connectionsInfo, _ := connectionsInfo.GetUdpInfo()
		for i, j := range connectionsInfo {
			fmt.Println("==========================================================")
			fmt.Println("PID:", i)
			for _, k := range j {
				fmt.Println("	Handle        :", k.Handle)
				fmt.Println("	Local Address :", k.LocalAddr)
				fmt.Println("	Remote Address:", k.RemoteAddr)
				fmt.Println("	------------------------------------------")
			}
		}

	}
}
