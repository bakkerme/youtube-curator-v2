package rss

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessRSSFeed_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		feedFile        string
		expectError     bool
		expectedTitle   string
		expectedEntries int
		validateFunc    func(t *testing.T, feed *Feed)
	}{
		{
			name:            "Valid Majuular Feed",
			feedFile:        "majuular_feed.xml",
			expectError:     false,
			expectedTitle:   "Majuular",
			expectedEntries: 2,
			validateFunc: func(t *testing.T, feed *Feed) {
				assert.Equal(t, "Majuular", feed.Title)
				assert.Equal(t, "https://www.youtube.com/channel/UCAYF6ZY9gWBR1GW3R7PX7yw", feed.URL.Href)
				assert.Len(t, feed.Entries, 2)

				// Validate first entry
				firstEntry := feed.Entries[0]
				assert.Equal(t, "yt:video:-krIPfk2OIA", firstEntry.ID)
				assert.Equal(t, "William Shatner's TekWar: A Forgotten Franchise in Retrospect", firstEntry.Title)
				assert.Equal(t, "https://www.youtube.com/watch?v=-krIPfk2OIA", firstEntry.Link.Href)
				assert.Equal(t, "https://i2.ytimg.com/vi/-krIPfk2OIA/hqdefault.jpg", firstEntry.MediaGroup.MediaThumbnail.URL)

				// Validate published time parsing
				expectedTime, _ := time.Parse(time.RFC3339, "2025-05-15T00:43:16+00:00")
				assert.Equal(t, expectedTime, firstEntry.Published)

				// Validate second entry
				secondEntry := feed.Entries[1]
				assert.Equal(t, "yt:video:0NyaGRNH2zE", secondEntry.ID)
				assert.Equal(t, "Ultima VII: The Black Gate Retrospective | Peak of the Golden Age", secondEntry.Title)
				assert.Equal(t, "https://i1.ytimg.com/vi/0NyaGRNH2zE/hqdefault.jpg", secondEntry.MediaGroup.MediaThumbnail.URL)
			},
		},
		{
			name:            "Empty Feed",
			feedFile:        "empty_feed.xml",
			expectError:     false,
			expectedTitle:   "Empty Test Channel",
			expectedEntries: 0,
			validateFunc: func(t *testing.T, feed *Feed) {
				assert.Equal(t, "Empty Test Channel", feed.Title)
				assert.Equal(t, "https://www.youtube.com/channel/UCEMPTYTEST", feed.URL.Href)
				assert.Len(t, feed.Entries, 0)
			},
		},
		{
			name:            "Malformed Feed",
			feedFile:        "malformed_feed.xml",
			expectError:     true,
			expectedTitle:   "",
			expectedEntries: 0,
			validateFunc:    nil, // No validation needed for error case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test feed file
			testDir := filepath.Join("test")
			filePath := filepath.Join(testDir, tt.feedFile)
			input, err := os.ReadFile(filePath)
			require.NoError(t, err, "Failed to read test feed file")

			// Process the RSS feed
			feed := &Feed{}
			err = processRSSFeed(string(input), feed)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				return
			}

			assert.NoError(t, err, "Unexpected error processing RSS feed")
			assert.NotEmpty(t, feed.RawRSS, "Raw RSS should be stored")
			assert.Equal(t, string(input), feed.RawRSS, "Raw RSS should match input")

			// Run custom validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, feed)
			}
		})
	}
}

func TestCleanContent_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		maxLen            int
		disableTruncation bool
		expected          string
	}{
		{
			name:              "Basic HTML stripping",
			input:             "<p>Hello <strong>world</strong>!</p>",
			maxLen:            100,
			disableTruncation: false,
			expected:          "Hello world!",
		},
		{
			name:              "HTML entities replacement",
			input:             "Don&#39;t forget &quot;quotes&quot; and&#32;spaces",
			maxLen:            100,
			disableTruncation: false,
			expected:          "Don't forget \"quotes\" and spaces",
		},
		{
			name:              "Truncation with ellipsis",
			input:             "This is a very long string that should be truncated",
			maxLen:            20,
			disableTruncation: false,
			expected:          "This is a very long ...",
		},
		{
			name:              "No truncation when disabled",
			input:             "This is a very long string that should not be truncated",
			maxLen:            20,
			disableTruncation: true,
			expected:          "This is a very long string that should not be truncated",
		},
		{
			name:              "String shorter than max length",
			input:             "Short text",
			maxLen:            100,
			disableTruncation: false,
			expected:          "Short text",
		},
		{
			name:              "Empty string",
			input:             "",
			maxLen:            100,
			disableTruncation: false,
			expected:          "",
		},
		{
			name:              "Complex HTML with entities",
			input:             "<div><p>The 1980&#39;s was a <em>glorious</em> time for sci-fi. With &quot;Cyberpunk&quot; at the cutting edge.</p></div>",
			maxLen:            50,
			disableTruncation: false,
			expected:          "The 1980's was a glorious time for sci-fi. With \"C...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanContent(tt.input, tt.maxLen, tt.disableTruncation)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFeed_FeedString(t *testing.T) {
	tests := []struct {
		name     string
		rawRSS   string
		expected string
	}{
		{
			name:     "Basic feed string",
			rawRSS:   "<feed><title>Test</title></feed>",
			expected: "<feed><title>Test</title></feed>",
		},
		{
			name:     "Empty feed string",
			rawRSS:   "",
			expected: "",
		},
		{
			name:     "Complex feed string",
			rawRSS:   "<?xml version=\"1.0\"?><feed xmlns=\"http://www.w3.org/2005/Atom\"><title>Complex Feed</title></feed>",
			expected: "<?xml version=\"1.0\"?><feed xmlns=\"http://www.w3.org/2005/Atom\"><title>Complex Feed</title></feed>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed := &Feed{RawRSS: tt.rawRSS}
			result := feed.FeedString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEntryString_TableDriven(t *testing.T) {
	publishedTime, _ := time.Parse(time.RFC3339, "2025-01-15T12:00:00Z")

	tests := []struct {
		name     string
		entry    Entry
		contains []string // Strings that should be present in the output
	}{
		{
			name: "Complete entry",
			entry: Entry{
				Title:      "Test Video Title",
				ID:         "yt:video:TEST123",
				Link:       Link{Href: "https://www.youtube.com/watch?v=TEST123"},
				Published:  publishedTime,
				Content:    "<p>This is test content with <strong>HTML</strong> tags.</p>",
				MediaGroup: MediaGroup{MediaThumbnail: MediaThumbnail{URL: "https://example.com/thumb.jpg"}},
			},
			contains: []string{
				"Title: Test Video Title",
				"ID: yt:video:TEST123",
				"Link: https://www.youtube.com/watch?v=TEST123",
				"Thumbnail: https://example.com/thumb.jpg",
				"This is test content with HTML tags",
			},
		},
		{
			name: "Entry with empty fields",
			entry: Entry{
				Title: "  Trimmed Title  ",
				ID:    "",
				Link:  Link{Href: ""},
			},
			contains: []string{
				"Title: Trimmed Title",
				"ID: ",
				"Link: ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.String()
			for _, expectedSubstring := range tt.contains {
				assert.Contains(t, result, expectedSubstring)
			}
		})
	}
}

func TestEntry_GetMethods(t *testing.T) {
	entry := Entry{
		ID:      "test-id-123",
		Content: "test content here",
	}

	assert.Equal(t, "test-id-123", entry.GetID())
	assert.Equal(t, "test content here", entry.GetContent())
}

func TestEntry_UnmarshalXML_TimeFormats(t *testing.T) {
	tests := []struct {
		name         string
		xmlInput     string
		expectError  bool
		expectedTime time.Time
	}{
		{
			name: "RFC3339 format",
			xmlInput: `<entry xmlns="http://www.w3.org/2005/Atom">
				<published>2025-01-15T12:00:00Z</published>
			</entry>`,
			expectError:  false,
			expectedTime: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "RFC3339 with timezone",
			xmlInput: `<entry xmlns="http://www.w3.org/2005/Atom">
				<published>2025-01-15T12:00:00+05:00</published>
			</entry>`,
			expectError: false,
			expectedTime: func() time.Time {
				t, _ := time.Parse(time.RFC3339, "2025-01-15T12:00:00+05:00")
				return t
			}(),
		},
		{
			name: "Invalid format",
			xmlInput: `<entry xmlns="http://www.w3.org/2005/Atom">
				<published>invalid-date</published>
			</entry>`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entry Entry
			err := xml.Unmarshal([]byte(tt.xmlInput), &entry)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error unmarshaling XML")
				if !tt.expectedTime.IsZero() {
					assert.Equal(t, tt.expectedTime, entry.Published)
				}
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkProcessRSSFeed(b *testing.B) {
	testDir := filepath.Join("test")
	filePath := filepath.Join(testDir, "majuular_feed.xml")
	input, err := os.ReadFile(filePath)
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	inputStr := string(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed := &Feed{}
		_ = processRSSFeed(inputStr, feed)
	}
}

func BenchmarkCleanContent(b *testing.B) {
	input := "<p>The 1980&#39;s was a <em>glorious</em> time for sci-fi. With &quot;Cyberpunk&quot; at the cutting edge of fiction and Star Trek&#39;s popularity resurging thanks to The Next Generation, the beloved Captain Kirk himself dipped his feet into the world of pulp literature.</p>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CleanContent(input, 200, false)
	}
}
