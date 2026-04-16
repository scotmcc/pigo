package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/scotmcc/pigo/internal/config"
)

// pidFilePath is ~/.pigo/pigo.pid — sits alongside config.toml and pigo.db.
func pidFilePath(cfg config.Config) string {
	return filepath.Join(filepath.Dir(cfg.Vault.Path), "pigo.pid")
}

// writePidFile records the current process's PID so external tools (and
// `pigo stop` / `pigo status`) can find the daemon.
func writePidFile(cfg config.Config) error {
	path := pidFilePath(cfg)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create pid file directory: %w", err)
	}
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0644)
}

// readPidFile returns the PID recorded in the file. Returns an error wrapping
// os.ErrNotExist when the daemon isn't running (or never wrote a pid file).
func readPidFile(cfg config.Config) (int, error) {
	data, err := os.ReadFile(pidFilePath(cfg))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse pid file: %w", err)
	}
	return pid, nil
}

// removePidFile removes the pid file. Silent no-op when it doesn't exist.
func removePidFile(cfg config.Config) {
	os.Remove(pidFilePath(cfg))
}

// processAlive uses signal 0 — POSIX's "am I allowed to signal this?" probe —
// to check whether a PID refers to a live process without actually disturbing it.
func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
