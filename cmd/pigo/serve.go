package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scotmcc/pigo/internal/commands"
	"github.com/scotmcc/pigo/internal/search"
	"github.com/scotmcc/pigo/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the pigo server",
	Long:  "Start the pigo daemon. Accepts commands over HTTP and TCP pipe.",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	commands.SetVaultService(svc)

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Refuse to start if another pigo is already running — two daemons on the
	// same HTTP/pipe ports would collide confusingly.
	if existing, err := readPidFile(cfg); err == nil && processAlive(existing) {
		return fmt.Errorf("pigo is already running (pid %d) — use 'pigo stop' first", existing)
	}

	// Wire search client if configured.
	if cfg.Search.URL != "" {
		commands.SetSearchClient(search.NewClient(cfg.Search.URL))
	}

	// Start the HTTP server in a goroutine.
	httpSrv := server.New(cfg.Server.Host, cfg.Server.Port)
	go func() {
		if err := httpSrv.Start(); err != nil && err.Error() != "http: Server closed" {
			fmt.Fprintf(os.Stderr, "http server error: %v\n", err)
		}
	}()

	// Start the TCP pipe server in a goroutine.
	pipe := server.NewPipe(cfg.Server.Host, cfg.Server.PipePort)
	go func() {
		if err := pipe.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "pipe server error: %v\n", err)
		}
	}()

	// Record our PID so 'pigo stop' / 'pigo status' can find us.
	if err := writePidFile(cfg); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}
	defer removePidFile(cfg)

	fmt.Printf("pigo serving (pid %d) — HTTP %s:%d, pipe %s:%d\n",
		os.Getpid(), cfg.Server.Host, cfg.Server.Port, cfg.Server.Host, cfg.Server.PipePort)

	// Wait for shutdown signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	fmt.Printf("\nreceived %s, shutting down...\n", sig)
	httpSrv.Shutdown(5 * time.Second)
	pipe.Close()
	fmt.Println("server stopped")

	return nil
}
