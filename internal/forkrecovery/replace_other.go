//go:build !windows

package forkrecovery

import "os"

func replaceFile(source, destination string) error {
	return os.Rename(source, destination)
}
