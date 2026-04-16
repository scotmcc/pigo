package vault

import "testing"

func TestDetectWikiLinks(t *testing.T) {
	body := `This relates to [[why-we-chose-go]] and also [[sqlite-patterns]].
See [[why-we-chose-go]] again for details.
Also check [[Some Title With Spaces|display text]].`

	links := DetectWikiLinks(body)

	if len(links) != 3 {
		t.Fatalf("expected 3 unique links, got %d: %v", len(links), links)
	}

	want := map[string]bool{
		"why-we-chose-go":          true,
		"sqlite-patterns":          true,
		"some-title-with-spaces":   true,
	}
	for _, link := range links {
		if !want[link] {
			t.Errorf("unexpected link: %q", link)
		}
	}
}

func TestDetectWikiLinks_None(t *testing.T) {
	links := DetectWikiLinks("Just plain text, no links.")
	if len(links) != 0 {
		t.Errorf("expected no links, got %v", links)
	}
}

func TestDetectWikiLinks_Empty(t *testing.T) {
	links := DetectWikiLinks("")
	if len(links) != 0 {
		t.Errorf("expected no links, got %v", links)
	}
}
