package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// printJSON writes v as indented JSON to stdout.
// Used when --json flag is set.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
