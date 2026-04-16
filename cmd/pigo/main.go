// pigo — a resident AI knowledge server.
//
// This is the entry point. It builds the root command and runs it.
// All actual logic lives in internal/ packages.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
