// Package git wraps go-git for vault version control.
// Layer-1 package — talks to git and nothing else.
package git

import (
	"fmt"
	"os"

	gogit "github.com/go-git/go-git/v5"
)

// Repo wraps a go-git repository.
type Repo struct {
	repo *gogit.Repository
	path string
}

// Open opens an existing git repo, or initializes a new one if it doesn't exist.
func Open(path string) (*Repo, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("create repo directory: %w", err)
	}

	repo, err := gogit.PlainOpen(path)
	if err == gogit.ErrRepositoryNotExists {
		repo, err = gogit.PlainInit(path, false)
		if err != nil {
			return nil, fmt.Errorf("init repo: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	return &Repo{repo: repo, path: path}, nil
}
