// +build windows

package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

func HideFile(filename string) (string, error) {
	filenameW, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return "", err
	}
	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func DiskStatus(disk string) (DiskUsage, error) {
	kernel32, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return DiskUsage{}, err
	}
	defer func() {
		_ = syscall.FreeLibrary(kernel32)
	}()

	GetDiskFreeSpaceEx, err := syscall.GetProcAddress(kernel32, "GetDiskFreeSpaceExW")
	if err != nil {
		return DiskUsage{}, err
	}

	diskNamePtr, err := syscall.UTF16PtrFromString(disk)
	if err != nil {
		return DiskUsage{}, err
	}

	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	_, _, e := syscall.Syscall6(GetDiskFreeSpaceEx, 4,
		uintptr(unsafe.Pointer(diskNamePtr)),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)), 0, 0)
	if e != 0 {
		return DiskUsage{}, errors.New("failed to load disk status")
	}

	all := uint64(lpTotalNumberOfBytes)
	free := uint64(lpTotalNumberOfFreeBytes)

	ds := DiskUsage{
		All:  all,
		Free: free,
		Used: all - free,
	}

	/*logs.Printf("Available  %dmb", lpFreeBytesAvailable/1024/1024.0)
	logs.Printf("Total      %dmb", lpTotalNumberOfBytes/1024/1024.0)
	logs.Printf("Free       %dmb", lpTotalNumberOfFreeBytes/1024/1024.0)*/
	return ds, nil
}

func DriveList() (drives []string) {

	checkDriveWithinTimeOut := func(d int32) bool {
		driveChan := make(chan string, 1)
		go func() {
			_, err := os.Open(string(d) + ":\\")
			if err == nil {
				driveChan <- fmt.Sprintf("%c", d)
			}
		}()

		select {
		case <-driveChan:
			return true
		case <-time.After(time.Millisecond * 10):
			return false
		}
	}

	for _, d := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if checkDriveWithinTimeOut(d) {
			drives = append(drives, fmt.Sprintf("%c", d))
		}
	}
	return
}
