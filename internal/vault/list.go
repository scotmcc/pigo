package vault

import (
	"encoding/json"

	"github.com/scotmcc/pigo/internal/db"
)

// ListItem is a summary of a note for listing.
type ListItem struct {
	ID        string
	Title     string
	Tags      []string
	UpdatedAt string
}

// List returns all notes in the vault as summaries.
func (s *Service) List() ([]ListItem, error) {
	notes, err := s.db.ListNotes()
	if err != nil {
		return nil, err
	}

	items := make([]ListItem, len(notes))
	for i, n := range notes {
		items[i] = noteToListItem(n)
	}
	return items, nil
}

func noteToListItem(n db.Note) ListItem {
	var tags []string
	json.Unmarshal([]byte(n.Tags), &tags)

	return ListItem{
		ID:        n.ID,
		Title:     n.Title,
		Tags:      tags,
		UpdatedAt: n.UpdatedAt.Format("2006-01-02"),
	}
}
