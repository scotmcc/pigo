package vault

import (
	"encoding/json"
	"sort"
)

// TagCount holds a tag and how many notes use it.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// Tags returns all tags used across the vault with their note counts, sorted by count descending.
func (s *Service) Tags() ([]TagCount, error) {
	notes, err := s.db.ListNotes()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, n := range notes {
		var tags []string
		json.Unmarshal([]byte(n.Tags), &tags)
		for _, t := range tags {
			counts[t]++
		}
	}

	result := make([]TagCount, 0, len(counts))
	for tag, count := range counts {
		result = append(result, TagCount{Tag: tag, Count: count})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result, nil
}
