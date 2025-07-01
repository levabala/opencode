package diff

import (
	"image/color"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/compat"
	"github.com/sst/opencode/internal/theme"
)

func setupTestTheme() {
	// Create a system theme for testing
	testTheme := theme.NewSystemTheme(color.RGBA{0, 0, 0, 255}, true)
	theme.RegisterTheme("test", testTheme)
	theme.SetTheme("test")
}

func TestApplyHighlighting_UTF8Characters(t *testing.T) {
	setupTestTheme()
	tests := []struct {
		name        string
		content     string
		segments    []Segment
		segmentType LineType
		expectUTF8  bool
	}{
		{
			name:        "ASCII characters only",
			content:     "hello world",
			segments:    []Segment{{Start: 0, End: 5, Type: LineAdded}},
			segmentType: LineAdded,
			expectUTF8:  true,
		},
		{
			name:        "UTF-8 emoji characters",
			content:     "hello 🌍 world",
			segments:    []Segment{{Start: 6, End: 7, Type: LineAdded}},
			segmentType: LineAdded,
			expectUTF8:  true,
		},
		{
			name:        "UTF-8 accented characters",
			content:     "café résumé naïve",
			segments:    []Segment{{Start: 0, End: 4, Type: LineAdded}},
			segmentType: LineAdded,
			expectUTF8:  true,
		},
		{
			name:        "UTF-8 Chinese characters",
			content:     "你好世界",
			segments:    []Segment{{Start: 0, End: 2, Type: LineAdded}},
			segmentType: LineAdded,
			expectUTF8:  true,
		},
		{
			name:        "UTF-8 mixed with ANSI codes",
			content:     "\x1b[31m你好\x1b[0m世界",
			segments:    []Segment{{Start: 0, End: 2, Type: LineAdded}},
			segmentType: LineAdded,
			expectUTF8:  true,
		},
		{
			name:        "UTF-8 mathematical symbols",
			content:     "∑∏∆√∞≠≤≥",
			segments:    []Segment{{Start: 2, End: 5, Type: LineAdded}},
			segmentType: LineAdded,
			expectUTF8:  true,
		},
		{
			name:        "UTF-8 Arabic text",
			content:     "مرحبا بالعالم",
			segments:    []Segment{{Start: 0, End: 5, Type: LineAdded}},
			segmentType: LineAdded,
			expectUTF8:  true,
		},
	}

	highlightBg := compat.AdaptiveColor{
		Light: lipgloss.Color("#ffff00"),
		Dark:  lipgloss.Color("#444400"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyHighlighting(tt.content, tt.segments, tt.segmentType, highlightBg)

			// Verify the result is valid UTF-8
			if !isValidUTF8(result) {
				t.Errorf("applyHighlighting() produced invalid UTF-8 for input %q", tt.content)
			}

			// Verify the original UTF-8 characters are preserved (ignoring ANSI codes)
			originalChars := extractVisibleChars(tt.content)
			resultChars := extractVisibleChars(result)

			if originalChars != resultChars {
				t.Errorf("applyHighlighting() corrupted UTF-8 characters:\noriginal: %q\nresult:   %q", originalChars, resultChars)
			}

			// Verify the result contains highlighting ANSI codes when segments are present
			if len(tt.segments) > 0 && !containsHighlightingCodes(result) {
				t.Errorf("applyHighlighting() did not add highlighting codes for segments")
			}
		})
	}
}

func TestApplyHighlighting_PreservesExistingANSI(t *testing.T) {
	setupTestTheme()
	content := "\x1b[31mhello\x1b[0m 🌍 \x1b[32mworld\x1b[0m"
	segments := []Segment{{Start: 6, End: 7, Type: LineAdded}}
	highlightBg := compat.AdaptiveColor{
		Light: lipgloss.Color("#ffff00"),
		Dark:  lipgloss.Color("#444400"),
	}

	result := applyHighlighting(content, segments, LineAdded, highlightBg)

	// Verify UTF-8 is preserved
	if !isValidUTF8(result) {
		t.Errorf("applyHighlighting() produced invalid UTF-8")
	}

	// Verify original visible characters are preserved
	originalChars := extractVisibleChars(content)
	resultChars := extractVisibleChars(result)

	if originalChars != resultChars {
		t.Errorf("applyHighlighting() corrupted characters with existing ANSI:\noriginal: %q\nresult:   %q", originalChars, resultChars)
	}
}

// isValidUTF8 checks if a string contains valid UTF-8 sequences
func isValidUTF8(s string) bool {
	return strings.ToValidUTF8(s, "") == s
}

// extractVisibleChars removes ANSI escape sequences and returns only visible characters
func extractVisibleChars(s string) string {
	// Remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b(?:[@-Z\\-_]|\[[0-9?]*(?:;[0-9?]*)*[@-~])`)
	return ansiRegex.ReplaceAllString(s, "")
}

// containsHighlightingCodes checks if the string contains color highlighting ANSI codes
func containsHighlightingCodes(s string) bool {
	// Look for 38;2 (foreground) or 48;2 (background) RGB color codes
	return strings.Contains(s, "\x1b[38;2;") || strings.Contains(s, "\x1b[48;2;")
}
