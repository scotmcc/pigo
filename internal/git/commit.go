package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitFile stages a file and commits it with the given message.
// The filePath is relative to the repo root.
func (r *Repo) CommitFile(filePath string, message string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("get worktree: %w", err)
	}

	if _, err := w.Add(filePath); err != nil {
		return fmt.Errorf("stage file %s: %w", filePath, err)
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "pigo",
			Email: "pigo@localhost",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
