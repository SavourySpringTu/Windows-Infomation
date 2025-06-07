package sysInfo

var (
	procGetTickCount64 = kernel32.NewProc("GetTickCount64")
)

func GetTimeUp() (int, error) {

	ret, _, err := procGetTickCount64.Call()

	result := int(ret) / 1000 / 60

	if ret == 0 {
		return result, err
	}
	return result, nil
}
