package files

import (
	"os"
	"runtime"
	"strings"
)

type DiskUsage struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

func NormalizePath(p string) string {
	if runtime.GOOS != "windows" {
		return p
	}
	drive := p[0:1]
	rest := p[2:]
	return "/" + strings.ToLower(drive) + strings.Replace(rest, "\\", "/", -1)
}

func UnNormalizePath(p string) string {
	if runtime.GOOS != "windows" {
		return p
	}
	drive := p[1:2]
	rest := p[3:]
	return strings.ToUpper(drive) + ":\\" + strings.Replace(rest, "/", "\\", -1)
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
