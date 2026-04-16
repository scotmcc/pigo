package main

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/scotmcc/pigo/internal/config"
)

// checkPortAvailable tries to bind host:port once and immediately closes. A
// bind failure returns a user-friendly error naming the offending process
// (via lsof) and pointing at the config file for remediation. The label
// distinguishes HTTP vs pipe ports in the error message.
func checkPortAvailable(cfg config.Config, port int, label string) error {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		holder := whoHasPort(port)
		configPath := filepath.Join(filepath.Dir(cfg.Vault.Path), "config.toml")
		if holder != "" {
			return fmt.Errorf(
				"%s port %d is in use by %s — stop it, or change [server] %s in %s",
				label, port, holder, configKeyForLabel(label), configPath,
			)
		}
		return fmt.Errorf(
			"%s port %d is in use — free the port, or change [server] %s in %s",
			label, port, configKeyForLabel(label), configPath,
		)
	}
	ln.Close()
	return nil
}

// whoHasPort shells out to lsof to identify the process holding a TCP port.
// Returns a short "<command> (pid <pid>)" string, or empty when lsof isn't
// available or the port lookup turns up nothing.
func whoHasPort(port int) string {
	out, err := exec.Command(
		"lsof", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-n", "-P",
	).Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return ""
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return ""
	}
	return fmt.Sprintf("%s (pid %s)", fields[0], fields[1])
}

func configKeyForLabel(label string) string {
	if label == "pipe" {
		return "pipe_port"
	}
	return "port"
}
