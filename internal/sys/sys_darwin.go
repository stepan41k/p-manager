//go:build darwin

package sys

import "golang.org/x/sys/unix"

func DisableMemoryDumps() {
	_ = unix.PtraceDenyAttach()
}