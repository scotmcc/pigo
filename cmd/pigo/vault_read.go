package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var vaultReadCmd = &cobra.Command{
	Use:   "read [id]",
	Short: "Read a note by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runVaultRead,
}

func init() {
	vaultCmd.AddCommand(vaultReadCmd)
}

func runVaultRead(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	result, err := svc.Read(args[0])
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(result)
	}

	fmt.Println(result.RawContent)
	return nil
}
