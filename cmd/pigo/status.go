package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether the pigo daemon is running",
	Long:  "Reports on the daemon: pid file, live process check, HTTP health ping. Exits 0 when running and reachable, 1 otherwise (useful in scripts).",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pid, err := readPidFile(cfg)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("pigo is not running (no pid file)")
			os.Exit(1)
		}
		return err
	}

	if !processAlive(pid) {
		fmt.Printf("pigo is not running (stale pid file references pid %d)\n", pid)
		os.Exit(1)
	}

	url := fmt.Sprintf("http://%s:%d/health", cfg.Server.Host, cfg.Server.Port)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("pigo is running (pid %d) but HTTP unreachable at %s: %v\n", pid, url, err)
		os.Exit(1)
	}
	resp.Body.Close()

	fmt.Printf("pigo is running (pid %d) — HTTP %s, pipe %s:%d\n",
		pid, url, cfg.Server.Host, cfg.Server.PipePort)
	return nil
}
