package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running pigo daemon",
	Long:  "Reads ~/.pigo/pigo.pid and sends SIGTERM. Cleans up a stale pid file if the process is already gone.",
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pid, err := readPidFile(cfg)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("pigo is not running (no pid file)")
			return nil
		}
		return err
	}

	if !processAlive(pid) {
		fmt.Printf("pid file references dead process (pid %d) — cleaning up\n", pid)
		removePidFile(cfg)
		return nil
	}

	proc, _ := os.FindProcess(pid)
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM to pid %d: %w", pid, err)
	}

	fmt.Printf("sent SIGTERM to pigo (pid %d), waiting for shutdown...\n", pid)
	// Poll for exit — pigo's graceful shutdown has a 5s HTTP drain, so allow ~10s.
	for range 20 {
		if !processAlive(pid) {
			fmt.Println("pigo stopped")
			removePidFile(cfg)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("pigo didn't exit within 10s — check 'pigo status' or force with 'kill -9 %d'", pid)
}
