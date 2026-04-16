package vault

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Why We Chose Go", "why-we-chose-go"},
		{"SQLite as the Universal Database", "sqlite-as-the-universal-database"},
		{"Hello, World!", "hello-world"},
		{"  spaces  everywhere  ", "spaces-everywhere"},
		{"already-a-slug", "already-a-slug"},
		{"UPPERCASE", "uppercase"},
		{"special!@#$%chars", "special-chars"},
		{"multiple---hyphens", "multiple-hyphens"},
		{"", ""},
	}

	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
