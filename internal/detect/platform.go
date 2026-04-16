package detect

import (
	"os/exec"
	"runtime"
)

// OnPath reports whether the named binary is available on PATH.
func OnPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// IsMac reports whether the current OS is macOS.
func IsMac() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux reports whether the current OS is Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
