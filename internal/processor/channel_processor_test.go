package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"youtube-curator-v2/internal/rss"
)

// MockStore is a mock implementation of store.Store for testing
type MockStore struct {
	lastCheckedTimestamps map[string]time.Time
	lastCheckedVideoIDs   map[string]string
	getError              error
	setError              error
}

func NewMockStore() *MockStore {
	return &MockStore{
		lastCheckedTimestamps: make(map[string]time.Time),
		lastCheckedVideoIDs:   make(map[string]string),
	}
}

func (m *MockStore) GetLastCheckedTimestamp(channelID string) (time.Time, error) {
	if m.getError != nil {
		return time.Time{}, m.getError
	}
	return m.lastCheckedTimestamps[channelID], nil
}

func (m *MockStore) SetLastCheckedTimestamp(channelID string, timestamp time.Time) error {
	if m.setError != nil {
		return m.setError
	}
	m.lastCheckedTimestamps[channelID] = timestamp
	return nil
}

func (m *MockStore) GetLastCheckedVideoID(channelID string) (string, error) {
	if m.getError != nil {
		return "", m.getError
	}
	return m.lastCheckedVideoIDs[channelID], nil
}

func (m *MockStore) SetLastCheckedVideoID(channelID, videoID string) error {
	if m.setError != nil {
		return m.setError
	}
	m.lastCheckedVideoIDs[channelID] = videoID
	return nil
}

func (m *MockStore) Close() error {
	return nil
}

// MockFeedProvider is a mock implementation of rss.FeedProvider for testing
type MockFeedProvider struct {
	feeds map[string]*rss.Feed
	err   error
}

func NewMockFeedProvider() *MockFeedProvider {
	return &MockFeedProvider{
		feeds: make(map[string]*rss.Feed),
	}
}

func (m *MockFeedProvider) FetchFeed(ctx context.Context, channelID string) (*rss.Feed, error) {
	if m.err != nil {
		return nil, m.err
	}
	feed, ok := m.feeds[channelID]
	if !ok {
		return nil, errors.New("feed not found")
	}
	return feed, nil
}

func TestProcessChannel_NoNewVideos(t *testing.T) {
	// Setup
	mockStore := NewMockStore()
	mockFeedProvider := NewMockFeedProvider()
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider)

	channelID := "test-channel-1"
	lastChecked := time.Now().Add(-24 * time.Hour)
	mockStore.lastCheckedTimestamps[channelID] = lastChecked

	// Create a feed with videos older than last checked
	mockFeedProvider.feeds[channelID] = &rss.Feed{
		Entries: []rss.Entry{
			{
				Title:     "Old Video 1",
				Published: lastChecked.Add(-48 * time.Hour),
			},
			{
				Title:     "Old Video 2",
				Published: lastChecked.Add(-72 * time.Hour),
			},
		},
	}

	// Execute
	result := processor.ProcessChannel(context.Background(), channelID)

	// Verify
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	if result.NewVideo != nil {
		t.Errorf("Expected no new video, but got: %v", result.NewVideo)
	}
	if result.ChannelID != channelID {
		t.Errorf("Expected channel ID %s, but got %s", channelID, result.ChannelID)
	}
}

func TestProcessChannel_WithNewVideos(t *testing.T) {
	// Setup
	mockStore := NewMockStore()
	mockFeedProvider := NewMockFeedProvider()
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider)

	channelID := "test-channel-2"
	lastChecked := time.Now().Add(-24 * time.Hour)
	mockStore.lastCheckedTimestamps[channelID] = lastChecked

	// Create a feed with new videos
	newestVideoTime := time.Now().Add(-1 * time.Hour)
	mockFeedProvider.feeds[channelID] = &rss.Feed{
		Entries: []rss.Entry{
			{
				Title:     "Newest Video",
				Published: newestVideoTime,
				ID:        "newest-id",
			},
			{
				Title:     "Older New Video",
				Published: time.Now().Add(-2 * time.Hour),
				ID:        "older-new-id",
			},
			{
				Title:     "Old Video",
				Published: lastChecked.Add(-48 * time.Hour),
				ID:        "old-id",
			},
		},
	}

	// Execute
	result := processor.ProcessChannel(context.Background(), channelID)

	// Verify
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	if result.NewVideo == nil {
		t.Fatal("Expected a new video, but got nil")
	}
	if result.NewVideo.Title != "Newest Video" {
		t.Errorf("Expected newest video title 'Newest Video', but got '%s'", result.NewVideo.Title)
	}

	// Verify timestamp was updated
	storedTimestamp, _ := mockStore.GetLastCheckedTimestamp(channelID)
	if !storedTimestamp.Equal(newestVideoTime) {
		t.Errorf("Expected timestamp to be updated to %v, but got %v", newestVideoTime, storedTimestamp)
	}
}

func TestProcessChannel_FeedProviderError(t *testing.T) {
	// Setup
	mockStore := NewMockStore()
	mockFeedProvider := NewMockFeedProvider()
	mockFeedProvider.err = errors.New("network error")
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider)

	channelID := "test-channel-3"

	// Execute
	result := processor.ProcessChannel(context.Background(), channelID)

	// Verify
	if result.Error == nil {
		t.Fatal("Expected an error, but got nil")
	}
	if result.NewVideo != nil {
		t.Errorf("Expected no new video on error, but got: %v", result.NewVideo)
	}
}

func TestProcessChannel_FirstTimeCheck(t *testing.T) {
	// Setup - channel has never been checked before
	mockStore := NewMockStore()
	mockFeedProvider := NewMockFeedProvider()
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider)

	channelID := "new-channel"
	// Don't set any last checked timestamp - simulating first time

	// Create a feed with videos
	newestVideoTime := time.Now().Add(-1 * time.Hour)
	mockFeedProvider.feeds[channelID] = &rss.Feed{
		Entries: []rss.Entry{
			{
				Title:     "Recent Video",
				Published: newestVideoTime,
				ID:        "recent-id",
			},
			{
				Title:     "Older Video",
				Published: time.Now().Add(-24 * time.Hour),
				ID:        "older-id",
			},
		},
	}

	// Execute
	result := processor.ProcessChannel(context.Background(), channelID)

	// Verify
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	if result.NewVideo == nil {
		t.Fatal("Expected a new video for first-time check, but got nil")
	}
	if result.NewVideo.Title != "Recent Video" {
		t.Errorf("Expected newest video title 'Recent Video', but got '%s'", result.NewVideo.Title)
	}

	// Verify timestamp was set
	storedTimestamp, _ := mockStore.GetLastCheckedTimestamp(channelID)
	if !storedTimestamp.Equal(newestVideoTime) {
		t.Errorf("Expected timestamp to be set to %v, but got %v", newestVideoTime, storedTimestamp)
	}
}
