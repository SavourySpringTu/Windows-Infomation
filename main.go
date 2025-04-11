package main

import (
	"fmt"
	"main/sysinfo"
)

func main() {
	username, _ := sysinfo.GetUsername()
	fmt.Println(username)

	arch, _ := sysinfo.GetOSInfo()
	fmt.Println(arch)

	arch2, _ := sysinfo.GetOSInfo2()
	fmt.Println(arch2)

	timeup, _ := sysinfo.GetUptime()
	fmt.Println(timeup)

	sysinfo.GetAppInfo()

	//sysinfo.GetNameOS()
}
