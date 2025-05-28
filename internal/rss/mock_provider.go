package rss

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// MockFeedProvider implements the FeedProvider interface and provides mock feed data.
type MockFeedProvider struct {
	DefaultProvider FeedProvider
}

// NewMockFeedProvider creates a new instance of the mock feed provider.
func NewMockFeedProvider(defaultProvider FeedProvider) *MockFeedProvider {
	return &MockFeedProvider{
		DefaultProvider: defaultProvider,
	}
}

// FetchFeed retrieves a feed, either from a mock file or by fetching and storing it.
func (m *MockFeedProvider) FetchFeed(ctx context.Context, channelID string) (*Feed, error) {
	// Construct the mock file path based on the channel ID
	mockFilePath := filepath.Join("feed_mocks", fmt.Sprintf("%s.xml", channelID))

	// Check if the mock file exists
	_, err := os.Stat(mockFilePath)
	if err == nil {
		// File exists, read and return the mock data
		fmt.Printf("Loading mock feed from: %s\n", mockFilePath)
		b, err := os.ReadFile(mockFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read mock feed file %s: %w", mockFilePath, err)
		}

		feed := &Feed{}
		err = processRSSFeed(string(b), feed)
		if err != nil {
			return nil, fmt.Errorf("failed to process mock feed from %s: %w", mockFilePath, err)
		}
		return feed, nil
	} else if !os.IsNotExist(err) {
		// An error other than file not found occurred
		return nil, fmt.Errorf("error checking for mock feed file %s: %w", mockFilePath, err)
	}

	// File does not exist, fetch using the default provider
	fmt.Printf("Mock feed not found for %s, fetching and storing...\n", channelID)
	feed, err := m.DefaultProvider.FetchFeed(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed with default provider: %w", err)
	}

	// Store the fetched feed to a mock file for future use
	// Ensure the directory exists
	dir := filepath.Dir(mockFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create mock feed directory %s: %w", dir, err)
	}

	// Write the raw RSS content to the file
	if feed.RawRSS == "" {
		return nil, fmt.Errorf("fetched feed for %s has no raw RSS content to store", channelID)
	}
	if err := os.WriteFile(mockFilePath, []byte(feed.RawRSS), 0644); err != nil {
		return nil, fmt.Errorf("failed to write mock feed file %s: %w", mockFilePath, err)
	}

	fmt.Printf("Stored fetched feed to: %s\n", mockFilePath)

	return feed, nil
}
