package main

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var (
// sysInfoFlag = flag.Bool("sysinfo", false, "Get information of system")
// appInfoFlag = flag.Bool("appinfo", false, "Get information of all applications installed in the os")
// viewFlag    = flag.Bool("view", false, "Show GUI")
)

func main() {
	// flag.Parse()
	// if *sysInfoFlag {
	// 	username, _ := sysInfo.GetUsername()
	// 	fmt.Printf("User name: %s\n", username)

	// 	nameOS, _ := sysInfo.GetInfoOSbyName("ProductName")
	// 	fmt.Printf("Name OS: %s\n", nameOS)

	// 	versionOS, _ := sysInfo.GetInfoOSbyName("DisplayVersion")
	// 	fmt.Printf("Version OS: %s\n", versionOS)

	// 	buildOS, _ := sysInfo.GetInfoOSbyName("CurrentBuild")
	// 	fmt.Printf("Build OS: %s\n", buildOS)

	// 	timeup, _ := sysInfo.GetUptime()
	// 	fmt.Printf("Time up: %d minute\n", timeup)
	// }
	// if *appInfoFlag {
	// 	appInfo, _ := appInfo.GetAllAppInfo()
	// 	for _, item := range appInfo {
	// 		fmt.Printf("================ %s =================\n", item.Name)
	// 		fmt.Printf("Version: %s\n", item.Version)
	// 		fmt.Printf("Publisher: %s\n", item.Publisher)
	// 		fmt.Printf("Install Date: %s\n", item.InstallDate)
	// 	}
	// }

	var inTE, outTE *walk.TextEdit

	// Log xem chương trình đã bắt đầu
	fmt.Println("Starting GUI application...")

	MainWindow{
		Title:   "Hello Walk",
		MinSize: Size{300, 200},
		Layout:  VBox{},
		Children: []Widget{
			TextEdit{AssignTo: &inTE},
			PushButton{
				Text: "Greet",
				OnClicked: func() {
					outTE.SetText("Hello, " + inTE.Text())
				},
			},
			TextEdit{AssignTo: &outTE, ReadOnly: true},
		},
	}.Run()

	// Log sau khi chạy
	fmt.Println("GUI window should be visible now")
}
