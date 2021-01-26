// +build !windows

package files

import (
	//"github.com/jaypipes/ghw"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func HideFile(filename string) (string, error) {
	if !strings.HasPrefix(filepath.Base(filename), ".") {
		newPath := filepath.Join(filepath.Dir(filename), "."+filepath.Base(filename))
		err := os.Rename(filename, newPath)
		return newPath, err
	}
	return filename, nil
}

func DiskStatus(path string) (disk DiskUsage, err error) {
	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

/* func DriveList() []string {
	var drives []string

	block, err := ghw.Block()
	if err != nil {
		fmt.Printf("Error getting block storage info: %v", err)
	}

	for _, disk := range block.Disks {
		drives = append(drives, disk.Name)
	}
	return drives
} */
