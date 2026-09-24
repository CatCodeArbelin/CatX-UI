//go:build !windows

package analytics

import (
	"fmt"
	"os"
	"syscall"
)

func fileIdentity(info os.FileInfo) string {
	if data, ok := info.Sys().(*syscall.Stat_t); ok {
		return fmt.Sprintf("device=%d:inode=%d", data.Dev, data.Ino)
	}
	return info.Name()
}

func handleIdentity(_ *os.File, info os.FileInfo) string { return fileIdentity(info) }
