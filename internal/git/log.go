package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// LogEntry represents a single commit in the history.
type LogEntry struct {
	Hash    string
	Message string
	When    time.Time
}

// Log returns commit history for the entire repo, newest first.
// If limit is 0, returns all commits.
func (r *Repo) Log(limit int) ([]LogEntry, error) {
	iter, err := r.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("get log: %w", err)
	}
	defer iter.Close()

	var entries []LogEntry
	err = iter.ForEach(func(c *object.Commit) error {
		if limit > 0 && len(entries) >= limit {
			return fmt.Errorf("limit reached")
		}
		entries = append(entries, LogEntry{
			Hash:    c.Hash.String()[:8],
			Message: c.Message,
			When:    c.Author.When,
		})
		return nil
	})

	// "limit reached" is our stop signal, not a real error.
	if err != nil && err.Error() != "limit reached" {
		return nil, fmt.Errorf("iterate log: %w", err)
	}

	return entries, nil
}
