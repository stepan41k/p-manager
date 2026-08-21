//go:build linux

package sys

import "golang.org/x/sys/unix"

func DisableMemoryDumps() {
	_ = unix.Prctl(unix.PR_SET_DUMPABLE, 0, 0, 0, 0)
}