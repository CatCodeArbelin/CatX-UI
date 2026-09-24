//go:build windows

package analytics

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

func fileIdentity(info os.FileInfo) string {
	// Creation time is not unique at Windows timestamp resolution. The handle
	// file index and volume serial identify the actual file across renames.
	// The caller only has FileInfo, so this fallback is supplemented by the
	// stable creation time; openCurrent supplies the handle identity below.
	if data, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		return fmt.Sprintf("creation=%v", data.CreationTime)
	}
	return info.Name()
}

func handleIdentity(f *os.File, info os.FileInfo) string {
	var data windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &data); err == nil {
		return fmt.Sprintf("volume=%d:index=%d:%d", data.VolumeSerialNumber, data.FileIndexHigh, data.FileIndexLow)
	}
	return fileIdentity(info)
}
