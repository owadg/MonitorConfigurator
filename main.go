package main

import (
	"fmt"
	"syscall"
	"unsafe"
	/*
		"image/color"

		"fyne.io/fyne/v2"
		"fyne.io/fyne/v2/app"
		"fyne.io/fyne/v2/canvas"
		"fyne.io/fyne/v2/container"
		"fyne.io/fyne/v2/layout"
	*/)

const (
	DISPLAY_DEVICE_ATTACHED_TO_DESKTOP = 1
)

var (
	user32DLL                  = syscall.NewLazyDLL("user32.dll")
	procChangeDisplaySettingsA = user32DLL.NewProc("ChangeDisplaySettingsAs")
	procEnumDisplayDevicesA    = user32DLL.NewProc("EnumDisplayDevicesA") //params are (&Cstylestring, uint32, &Display_Device, uint32)
)

//some thoughts: The pointer to dispdeva really needs to be to 424 bytes of contiguous memory. currently no guarantees of that. will be good for parsing data, but thats it

//must set cb when calling EnumDispDevA
type DispDevA struct {
	cb           uint32 //size of struct should be 3376 bits, 424 bytes
	DeviceName   [32]uint8
	DeviceString [128]uint8
	StateFlags   uint32
	DeviceID     [128]uint8
	DeviceKey    [128]uint8
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

//this is just a simplified wrapper of the Windows method to get the display adapters
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

func dumpDispDev(dd *DispDevA) {
	fmt.Println("Size: ", dd.cb)
	fmt.Println("DeviceName: ", string(dd.DeviceName[0:]))
	fmt.Println("DeviceString: ", string(dd.DeviceString[0:]))
	fmt.Println("StateFlags: ", dd.StateFlags)
	fmt.Println("DeviceID: ", string(dd.DeviceID[0:]))
	fmt.Println("DeviceKey: ", string(dd.DeviceKey[0:]))
}

//returns an array containing all attached devices to any display adapter
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

//returns an array containing attached devices to a display adapter
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

//might just return a bunch of displays, or might queary all display adapters
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

func main() {
	/* BASE FUNCTIONALITY
	cb := DispDevA{}
	enumDispDev("", 0, &cb, 0x00000001)
	dumpDispDev(&cb)
	*/

	/*
		mons := shallowQueryDisplays()
		for i := range mons {
			dumpDispDev(&mons[i])
		}
	*/

	/*
		mons := getActiveMonitors()
		for i := range mons {
			dumpDispDev(&mons[i])
		}
	*/

}
