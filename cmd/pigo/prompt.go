package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// confirm prints a yes/no question and reads the user's answer from stdin.
// Empty input returns defaultYes. Returns false on read error so we never
// take action on input we couldn't actually read.
func confirm(question string, defaultYes bool) bool {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}
	fmt.Printf("%s %s ", question, hint)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer == "" {
		return defaultYes
	}
	return answer == "y" || answer == "yes"
}
