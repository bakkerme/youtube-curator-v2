package rss

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessRSSFeed_MockFiles(t *testing.T) {
	mockDir := filepath.Join("..", "..", "feed_mocks") // Adjust path as needed based on test file location

	// Test case 1: UCkpKS8M7MaZAFewtUz24K3A.xml
	t.Run("UCkpKS8M7MaZAFewtUz24K3A", func(t *testing.T) {
		filePath := filepath.Join(mockDir, "UCkpKS8M7MaZAFewtUz24K3A.xml")
		input, err := os.ReadFile(filePath)
		assert.NoError(t, err, "Failed to read mock file")

		feed := &Feed{}
		err = processRSSFeed(string(input), feed)
		assert.NoError(t, err, "Failed to process RSS feed")

		// Basic assertions - add more specific checks based on the file content
		assert.NotEmpty(t, feed.Title)
		assert.NotEmpty(t, feed.URL.Href)
		assert.Greater(t, len(feed.Entries), 0)
		assert.NotEmpty(t, feed.Entries[0].Title)
		assert.NotEmpty(t, feed.Entries[0].Link.Href)
		assert.NotEmpty(t, feed.Entries[0].ID)

		// Add more specific assertions for UCkpKS8M7MaZAFewtUz24K3A.xml if needed
	})

	// Test case 2: UCAYF6ZY9gWBR1GW3R7PX7yw.xml
	t.Run("UCAYF6ZY9gWBR1GW3R7PX7yw", func(t *testing.T) {
		filePath := filepath.Join(mockDir, "UCAYF6ZY9gWBR1GW3R7PX7yw.xml")
		input, err := os.ReadFile(filePath)
		assert.NoError(t, err, "Failed to read mock file")

		feed := &Feed{}
		err = processRSSFeed(string(input), feed)
		assert.NoError(t, err, "Failed to process RSS feed")

		// Basic assertions - add more specific checks based on the file content
		assert.NotEmpty(t, feed.Title)
		assert.NotEmpty(t, feed.URL.Href)
		assert.Greater(t, len(feed.Entries), 0)
		assert.NotEmpty(t, feed.Entries[0].Title)
		assert.NotEmpty(t, feed.Entries[0].Link.Href)
		assert.NotEmpty(t, feed.Entries[0].ID)

		// Add more specific assertions for UCAYF6ZY9gWBR1GW3R7PX7yw.xml if needed
	})
}
