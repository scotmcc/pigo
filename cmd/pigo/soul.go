package main

import (
	"fmt"

	"github.com/scotmcc/pigo/internal/assets"
	"github.com/spf13/cobra"
)

var soulCmd = &cobra.Command{
	Use:   "soul",
	Short: "View or initialize the pigo soul",
	Long: `The soul is your identity file — it tells the AI who you are, how you work,
and what you care about. It's stored as a vault note and injected into
every AI session.

Run 'pigo soul' to view it, or 'pigo soul init' to start the welcome flow.`,
	RunE: runSoul,
}

var soulInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Start the welcome flow to create your soul",
	Long: `Outputs the welcome prompt. In an AI session, this guides the AI through
getting to know you. The AI will ask questions and save your soul file.

With --prompt, outputs just the welcome prompt (for piping to an AI).
Without --prompt, checks if a soul exists and gives guidance.`,
	RunE: runSoulInit,
}

var soulPromptFlag bool

func init() {
	soulInitCmd.Flags().BoolVar(&soulPromptFlag, "prompt", false, "output raw welcome prompt (for piping to AI)")
	soulCmd.AddCommand(soulInitCmd)
	rootCmd.AddCommand(soulCmd)
}

func runSoul(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	result, err := svc.Read("soul")
	if err != nil {
		fmt.Println("No soul file found.")
		fmt.Println()
		fmt.Println("The soul is your identity — it tells the AI who you are and how you work.")
		fmt.Println("It gets injected into every AI session automatically.")
		fmt.Println()
		fmt.Println("To create one:")
		fmt.Println("  In a pi or Claude session, the AI will guide you through it automatically.")
		fmt.Println("  Or run: pigo soul init --prompt | pigo vault write --title Soul --tags system,identity --stdin")
		fmt.Println("  Or create ~/.pigo/vault/soul.md manually.")
		return nil
	}

	if jsonFlag {
		return printJSON(result)
	}

	fmt.Println(result.RawContent)
	return nil
}

func runSoulInit(cmd *cobra.Command, args []string) error {
	if soulPromptFlag {
		fmt.Print(assets.WelcomePrompt)
		return nil
	}

	// Check if soul already exists.
	svc, cleanup, err := setupVault()
	if err != nil {
		return err
	}
	defer cleanup()

	existing, _ := svc.Read("soul")
	if existing != nil {
		fmt.Println("Soul already exists. View it with: pigo soul")
		fmt.Println("Edit it with: pigo vault edit soul")
		return nil
	}

	fmt.Println("No soul file yet.")
	fmt.Println()
	fmt.Println("The easiest way to create one is in an AI session —")
	fmt.Println("the AI will ask you about yourself and save the soul automatically.")
	fmt.Println()
	fmt.Println("If you're in a pi session, it happens on first connect.")
	fmt.Println("If you're using Claude Code, run: /pigo soul init")
	fmt.Println()
	fmt.Println("Or create it manually:")
	fmt.Println("  pigo vault write --title Soul --tags system,identity --body '## User\\n- Name: ...'")

	return nil
}
