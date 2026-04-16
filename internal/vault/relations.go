package vault

import (
	"fmt"
	"os"
)

// maxRelations is the maximum number of auto-discovered relationships per note.
const maxRelations = 5

// discoverRelationships finds notes similar to the given note and updates
// its relates_to field. This is the mechanism that builds the knowledge graph
// automatically — the AI doesn't need to manually link things.
//
// It works by searching the vault using the note's content. The top N results
// (excluding the note itself) become relates_to entries.
func (s *Service) discoverRelationships(noteID string) error {
	// Read the note we just wrote.
	note, err := s.Read(noteID)
	if err != nil {
		return err
	}

	// Search for similar notes using the note's body as the query.
	// We use a shortened version to keep the embedding focused.
	query := note.Body
	if len(query) > 500 {
		query = query[:500]
	}

	result, err := s.Search(query, maxRelations+1)
	if err != nil {
		return err
	}

	// Filter out self and collect related note IDs.
	var related []string
	for _, r := range result.Results {
		if r.NoteID == noteID {
			continue
		}
		related = append(related, r.NoteID)
		if len(related) >= maxRelations {
			break
		}
	}

	if len(related) == 0 {
		return nil // no relationships found
	}

	// Re-parse the frontmatter and update relates_to.
	fm, body, err := ParseFrontmatter(note.RawContent)
	if err != nil {
		return err
	}

	fm.RelatesTo = related
	if err := s.updateNoteFile(noteID, fm, body, fmt.Sprintf("vault: %s [relations]", fm.Title)); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "discovered %d relationships for %s\n", len(related), noteID)
	return nil
}

// Backlinks returns all notes that have noteID in their relates_to or links_to.
// This is a reverse lookup — "who points at me?"
func (s *Service) Backlinks(noteID string) ([]string, error) {
	notes, err := s.db.ListNotes()
	if err != nil {
		return nil, err
	}

	var backlinks []string
	for _, n := range notes {
		if n.ID == noteID {
			continue
		}

		// Read the note and check its frontmatter.
		result, err := s.Read(n.ID)
		if err != nil {
			continue
		}

		for _, rel := range result.RelatesTo {
			if rel == noteID {
				backlinks = append(backlinks, n.ID)
				break
			}
		}

		// Also check links_to.
		fm, _, _ := ParseFrontmatter(result.RawContent)
		for _, link := range fm.LinksTo {
			if link == noteID {
				backlinks = append(backlinks, n.ID)
				break
			}
		}
	}

	return backlinks, nil
}

// LinksInfo holds all connection data for a note.
type LinksInfo struct {
	NoteID    string   `json:"note_id"`
	RelatesTo []string `json:"relates_to"` // auto-discovered similar notes
	LinksTo   []string `json:"links_to"`   // explicit [[wiki-links]] in body
	Backlinks []string `json:"backlinks"`  // notes that point at this one
}

// Links returns all connections for a note — relates_to, links_to, and backlinks.
func (s *Service) Links(noteID string) (*LinksInfo, error) {
	note, err := s.Read(noteID)
	if err != nil {
		return nil, err
	}

	fm, _, err := ParseFrontmatter(note.RawContent)
	if err != nil {
		return nil, err
	}

	backlinks, err := s.Backlinks(noteID)
	if err != nil {
		return nil, err
	}

	return &LinksInfo{
		NoteID:    noteID,
		RelatesTo: fm.RelatesTo,
		LinksTo:   fm.LinksTo,
		Backlinks: backlinks,
	}, nil
}
