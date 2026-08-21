//go:build !linux && !darwin

package sys

func DisableMemoryDumps() {}