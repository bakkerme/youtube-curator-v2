package rss

import (
	"testing"
	"time"
)

func TestEntry_String(t *testing.T) {
	publishedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	entry := Entry{
		Title:     "Test Video Title",
		Link:      Link{Href: "http://example.com/video"},
		ID:        "testvideoid",
		Published: publishedTime,
		MediaGroup: MediaGroup{
			MediaThumbnail: MediaThumbnail{
				URL: "http://example.com/thumbnail.jpg",
			},
		},
		Content: "This is a test video description with <b>HTML</b>.",
	}

	expectedString := `Title: Test Video Title
ID: testvideoid
Link: http://example.com/video
Published: 2023-01-01 12:00:00 +0000 UTC
Thumbnail: http://example.com/thumbnail.jpg
Content: This is a test video description with HTML.
`

	if entry.String() != expectedString {
		t.Errorf("Expected string:\n%s\nGot:\n%s", expectedString, entry.String())
	}
}

func TestEntry_String_NoThumbnail(t *testing.T) {
	publishedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	entry := Entry{
		Title:     "Test Video Title No Thumb",
		Link:      Link{Href: "http://example.com/video2"},
		ID:        "testvideoid2",
		Published: publishedTime,
		Content:   "Another description.",
	}

	expectedString := `Title: Test Video Title No Thumb
ID: testvideoid2
Link: http://example.com/video2
Published: 2023-01-01 12:00:00 +0000 UTC
Thumbnail: 
Content: Another description.
`

	if entry.String() != expectedString {
		t.Errorf("Expected string:\n%s\nGot:\n%s", expectedString, entry.String())
	}
}
