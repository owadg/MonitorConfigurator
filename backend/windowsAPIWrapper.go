package WindowsAPIWrapper

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"sort"

	"image/color"
)

const (
	//dmDisplayOrientation consts (The degrees measurement is when rotate 90 degrees)
	DMDO_DEFAULT = 0
	DMDO_90      = 1
	DMDO_180     = 2
	DMDO_270     = 3

	//enum display device consts
	DISPLAY_DEVICE_ATTACHED_TO_DESKTOP = 1

	//enum display settings
	ENUM_CURRENT_SETTINGS = -1

	//changedispsettings consts
	CCHDEVICENAME = 32
	CCHFORMNAME   = 32
)

var (
	user32DLL                  = syscall.NewLazyDLL("user32.dll")
	procChangeDisplaySettingsA = user32DLL.NewProc("ChangeDisplaySettingsA")
	procEnumDisplayDevicesA    = user32DLL.NewProc("EnumDisplayDevicesA")  //params are (&Cstylestring, uint32, &Display_Device, uint32)
	procEnumDisplaySettingsA   = user32DLL.NewProc("EnumDisplaySettingsA") //params are &cstrign, uint32, &devmode
	procEnumDisplayMonitors    = user32DLL.NewProc("EnumDisplayMonitors")  //params are HDC, LPCRECT, MONITORENUMPROC, LPARAM
	procGetMonitorInfoA        = user32DLL.NewProc("GetMonitorInfoA")      //params are hmonitor, lpmonitorinfo
)

// has all the info we need to run the GUI application, including active monitors, their settings
// and all possible settings
type GUIInfo struct {
	possibleSettings []SettingConfigList
	currentSettings  []MonitorInfo
	moreCurrSettings []DevMode
}

type SettingConfigList struct {
	monitor      DispDevA
	settingsList []DevMode //list of configurations for the given monitor
}

type Monitors struct {
	hmonitor []syscall.Handle
	hdc      []syscall.Handle
	rect     []*Rect
}

type Rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

// must set cb when calling EnumDispDevA
type DispDevA struct {
	cb           uint32 //size of struct should be 3376 bits, 424 bytes
	DeviceName   [32]uint8
	DeviceString [128]uint8
	StateFlags   uint32
	DeviceID     [128]uint8
	DeviceKey    [128]uint8
}

func dumpDispDev(dd *DispDevA) {
	fmt.Println("Size: ", dd.cb)
	fmt.Println("DeviceName: ", string(dd.DeviceName[0:]))
	fmt.Println("DeviceString: ", string(dd.DeviceString[0:]))
	fmt.Println("StateFlags: ", dd.StateFlags)
	fmt.Println("DeviceID: ", string(dd.DeviceID[0:]))
	fmt.Println("DeviceKey: ", string(dd.DeviceKey[0:]))
}

// https://docs.microsoft.com/en-us/windows/win32/api/windef/ns-windef-pointl
type pointl struct {
	x int32
	y int32
}

// https://docs.microsoft.com/en-us/windows/win32/api/wingdi/ns-wingdi-devmodea
type dummyStructName2 struct {
	dmPosition           pointl
	dmDisplayOrientation uint32
	dmDisplayFixedOutput uint32
}

// https://docs.microsoft.com/en-us/windows/win32/api/wingdi/ns-wingdi-devmodea
// must set dmSize when calling ChangeDisplaySettingsA
// one thing i have not determined is how size of devmode will work
type DevMode struct {
	dmDeviceName    [CCHDEVICENAME]byte
	dmSpecVersion   uint16
	dmDriverVersion uint16
	dmSize          uint16
	dmDriverExtra   uint16
	dmFields        uint32
	//I have determined that since I will only be using this for display devices, I only need to include DUMMYSTRUCTNAME2 for this union here, so I will be doing that
	//TODO: DETERMINE WHETHER OR NOT SIZEOFUNION IS JUST THE LARGEST MEMBER
	dummyUnionName dummyStructName2
	//note, I have determined short to be a 2 byte type, so int16, relevant comment because we have some shorts (will comment which are shorts)
	dmColor            int16
	dmDuplex           int16
	dmYResolution      int16
	dmTTOption         int16
	dmCollate          int16
	dmFormName         [CCHFORMNAME]byte
	dmLogPixels        uint16
	dmBitsPerPel       uint32
	dmPelsWidth        uint32
	dmPelsHeight       uint32
	dummyUnionName2    uint32 //can be either dmDisplayFlags, or more likely, dmNup <-- change display settings uses this! //either way, uint32
	dmDisplayFrequency uint32
	//the rest of these do not have to be declared, this is why yhe size is neccessary
	dmICMMethod     uint32
	dmICMIntent     uint32
	dmMediaType     uint32
	dmDitherType    uint32
	dmReserved1     uint32
	dmReserved2     uint32
	dmPanningWidth  uint32
	dmPanningHeight uint32
}

type MonitorInfo struct {
	cbSize    uint32
	rcMonitor Rect
	rcWork    Rect
	dwFlags   uint32
}

func dumpDevMode(dm *DevMode) {
	//this only dumps "important values" lol
	fmt.Println("dmDeviceName: ", string(dm.dmDeviceName[0:]))
	fmt.Println("dmSize: ", dm.dmSize)
	fmt.Println("dmBitsPerPel: ", dm.dmBitsPerPel)
	fmt.Println("dmPelsWiddth: ", dm.dmPelsWidth)
	fmt.Println("dmPelsHeight: ", dm.dmPelsHeight)
	fmt.Println("dmDisplayFrequency: ", dm.dmDisplayFrequency)
	fmt.Println("dmDisplayOrientation: ", dm.dummyUnionName.dmDisplayOrientation)
	fmt.Println("dmPointl: ", dm.dummyUnionName.dmPosition)
}

// StringToCharPtr converts a Go string into pointer to a null-terminated cstring.
// This assumes the go string is already ANSI encoded.
func StringToCharPtr(str string) *uint8 {
	if str == "" {
		return nil
	}

	chars := append([]byte(str), 0) // null terminated
	return &chars[0]
}

// this is just a simplified wrapper of the Windows method to get the display adapters
func enumDispDev(lpDevice string, iDevNum uint32, lpDisplayDevice *DispDevA, dwFlags uint32) uintptr {
	lpDisplayDevice.cb = uint32(unsafe.Sizeof(*lpDisplayDevice))

	r1, _, err := procEnumDisplayDevicesA.Call(uintptr(unsafe.Pointer(StringToCharPtr(lpDevice))),
		uintptr(iDevNum),
		uintptr(unsafe.Pointer(lpDisplayDevice)),
		uintptr(dwFlags))

	if err != syscall.Errno(0) {
		fmt.Println("error: ", err)
	}

	return r1
}

// returns an array containing all attached devices to any display adapter
func queryDisplayAdapters() []DispDevA {
	result := make([]DispDevA, 0)

	//we need to iterate through all display adapters, and then iterate through all attached monitors
	//iterating over all display adapters

	//this loops will call until it fails. This means the last element will always be one in
	//which the call failed. So, we will just remove it
	var fail bool = false
	for x := 0; !fail; x++ {
		temp, num := queryMonAttToDispAdapters(uint32(x))
		if num != 0 {
			result = append(result, temp...)
		} else {
			fail = true
		}
	}
	return result
}

// returns an array containing attached devices to a display adapter
func queryMonAttToDispAdapters(iDevNum uint32) ([]DispDevA, uintptr) {
	result := make([]DispDevA, 0)

	cb := DispDevA{}
	if enumDispDev("", iDevNum, &cb, 0x00000001) == 0 {
		fmt.Println("This Display Adapter does not exist, probably")
		return result, 0 // recall, 0 is an error
	}
	//now we have the adapter name in cb.DeviceString, and can check out attached devices

	//this loops will call until it fails. This means the last element will always be one in
	//which the call failed. So, we will just remove it
	var fail bool = false
	for x := 0; !fail; x++ {
		temp := DispDevA{}
		num := enumDispDev(string(cb.DeviceString[0:]), uint32(x), &temp, 0x00000001)

		fmt.Println("How it's going: ", num, "Adapter", string(cb.DeviceString[0:]), "Index: ", x, "obj in question: ", temp)
		if num != 0 {
			result = append(result, temp)
		} else {
			fail = true
		}
	}

	return result, 5 //5 just needs to be a nonzero number
}

// might just return a bunch of displays, or might queary all display adapters
func shallowQueryDisplays() []DispDevA {
	result := make([]DispDevA, 0)
	var fail bool = false
	for x := 0; !fail; x++ {
		temp := DispDevA{}
		num := enumDispDev("", uint32(x), &temp, 0x00000001)
		if num != 0 {
			result = append(result, temp)
		} else {
			fail = true
		}
	}
	return result
}

func getActiveMonitors() []DispDevA {
	alldisps := shallowQueryDisplays()
	result := make([]DispDevA, 0)

	for i := range alldisps {
		if (alldisps[i].StateFlags & DISPLAY_DEVICE_ATTACHED_TO_DESKTOP) == DISPLAY_DEVICE_ATTACHED_TO_DESKTOP {
			result = append(result, alldisps[i])
		}
	}
	return result
}

// this func is a wrapper for EnumDisplaySettingsA windows api function.
// we use it to get the possible settings for a display
func enumDisplaySettings(lpszDeviceName string, iModeNum uint32, lpDevMode *DevMode) uintptr {
	lpDevMode.dmSize = uint16(unsafe.Sizeof(lpDevMode))

	r1, _, err := procEnumDisplaySettingsA.Call(uintptr(unsafe.Pointer(StringToCharPtr(lpszDeviceName))),
		uintptr(iModeNum),
		uintptr(unsafe.Pointer(lpDevMode)))
	if err != syscall.Errno(0) {
		fmt.Println("error: ", err)
	}
	return r1
}

// should get list of possible settings
// if I want to optimize this, I might create a smaller structure than Devmode to fill this
func getPossibleSettingsForMonitor(lpszDeviceName string) []DevMode {
	result := make([]DevMode, 0)

	var fail bool = false
	for x := 0; !fail; x++ {
		temp := DevMode{}
		num := enumDisplaySettings(lpszDeviceName, uint32(x), &temp)
		if num != 0 {
			result = append(result, temp)
		} else {
			fail = true
		}
	}
	return result
}

// @return each key is a monitor and then all its possible configurations as the value
func getSettingsConfigList(monitors *[]DispDevA) []SettingConfigList {
	result := make([]SettingConfigList, 0)
	for i := range *monitors {
		temp := SettingConfigList{(*monitors)[i], getPossibleSettingsForMonitor(string((*monitors)[i].DeviceName[0:]))}
		result = append(result, temp)
	}
	return result
}

func getCurrentSettingsForCurrentMonitors() []DevMode {
	monitors := getActiveMonitors()
	result := make([]DevMode, 0)
	for i := range monitors {
		temp := DevMode{}
		enumDisplaySettings(string(monitors[i].DeviceName[0:]), 0xFFFFFFFF, &temp)
		result = append(result, temp)
	}
	return result
}

// I can enumerate the monitors I need here.
// the callback function should be called per monitor
// it's a little janky, and not giving me position info, so we use GetMonitor now!
func enumDisplayMonitors() Monitors {
	monin := Monitors{make([]syscall.Handle, 0), make([]syscall.Handle, 0), make([]*Rect, 0)}
	monproccallback := syscall.NewCallback(monitorEnumProc)

	r1, _, err := procEnumDisplayMonitors.Call(uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(monproccallback)),
		uintptr(unsafe.Pointer(&monin)))
	if r1 == 0 {
		fmt.Println("Failed. R1: ", r1)
	}
	if err != syscall.Errno(0) {
		fmt.Println("error: ", err)
	}

	return monin
}

// callback function for enumDIsplayMonitors
func monitorEnumProc(hMonitor syscall.Handle, hdc syscall.Handle, rect *Rect, inform *Monitors) uintptr {
	inform.hdc = append(inform.hdc, hdc)
	inform.hmonitor = append(inform.hmonitor, hMonitor)
	inform.rect = append(inform.rect, rect)
	return 500 //random value lol
}

// wrapper for you know it, GetMonitorInfoA
// i do not know if i need to pass handles with unsafe.Pointer, but assuming i dont for right now
func GetMonitorInfo(hMonitor syscall.Handle) MonitorInfo {
	temp := MonitorInfo{}
	temp.cbSize = uint32(unsafe.Sizeof(temp))

	r1, _, err := procGetMonitorInfoA.Call(uintptr(hMonitor), uintptr(unsafe.Pointer(&temp)))
	if r1 == 0 {
		fmt.Println("Failed. R1: ", r1)
	}
	if err != syscall.Errno(0) {
		fmt.Println("error: ", err)
	}
	return temp
}

func GetAllMonitorInfo() []MonitorInfo {
	result := make([]MonitorInfo, 0)

	monitors := enumDisplayMonitors()
	for i := range monitors.hmonitor {
		result = append(result, GetMonitorInfo(monitors.hmonitor[i]))
	}

	return result
}

/*
		//BASE FUNCTIONALITY
		//basic test for calling enumDispDev
		cb := DispDevA{}
		enumDispDev("", 0, &cb, 0x00000001)
		dumpDispDev(&cb)
	*/

	/*
		//lists all display adapters
		mons := shallowQueryDisplays()
		for i := range mons {
			dumpDispDev(&mons[i])
		}
	*/

	/*
		//gets all active monitor
		mons := getActiveMonitors()
		for i := range mons {
			dumpDispDev(&mons[i])
		}
	*/

	/*
		//basic test for calling enumDispSettings
		monitors := getActiveMonitors()
		dm := DevMode{}
		fmt.Println(dm)
		r1 := enumDisplaySettings(string(monitors[0].DeviceName[0:]), 0xFFFFFFFF, &dm)
		fmt.Println("test, r1:", r1)
		fmt.Println(dm)
	*/

	/*
		//test for getting all settings of a monitor
		monitor := getActiveMonitors()
		settings := getPossibleSettingsForMonitor(string(monitor[0].DeviceName[0:]))
		for i := range settings {
			dumpDevMode(&settings[i])
		}
	*/

	/*
		//testing getting all possible settings for all monitors
		monitors := getActiveMonitors()
		list := getSettingsConfigList(&monitors)
		for i := range list[0].settingsList {
			fmt.Println(list[0].settingsList[i])
		}
	*/

	/*
		//simple test of enum display monitors.
		//TODO: fix the rectangle
		info := enumDisplayMonitors()
		fmt.Println(*(info.rect[0]))
	*/

	/*
		//test for display position. Note this is based on virtual coordinate
		fmt.Println(GetMonitorInfo(enumDisplayMonitors().hmonitor[0]).rcWork)
		fmt.Println(GetMonitorInfo(enumDisplayMonitors().hmonitor[1]).rcWork)
	*/

	/*
		//getting curr settings
		moreeSet := getCurrentSettingsForCurrentMonitors()
		for i := range moreeSet {
			dumpDevMode(&moreeSet[i])
		}
	*/