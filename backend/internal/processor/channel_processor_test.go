package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"go.uber.org/mock/gomock"
)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider)

	channelID := "test-channel-1"
	lastChecked := time.Now().Add(-24 * time.Hour)

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

	// Set up expectations for the mock
	mockStore.EXPECT().GetLastCheckedTimestamp(channelID).Return(lastChecked, nil)
	// SetLastCheckedTimestamp should not be called when there are no new videos

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider)

	channelID := "test-channel-2"
	lastChecked := time.Now().Add(-24 * time.Hour)

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

	// Set up expectations for the mock
	var capturedTimestamp time.Time
	mockStore.EXPECT().GetLastCheckedTimestamp(channelID).Return(lastChecked, nil)
	mockStore.EXPECT().SetLastCheckedTimestamp(channelID, gomock.Any()).Do(func(channelID string, timestamp time.Time) {
		capturedTimestamp = timestamp
	})

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

	// Verify timestamp was updated to the newest video time
	if !capturedTimestamp.Equal(newestVideoTime) {
		t.Errorf("Expected timestamp to be updated to %v, but got %v", newestVideoTime, capturedTimestamp)
	}
}

func TestProcessChannel_FeedProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	mockFeedProvider.err = errors.New("network error")
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider)

	channelID := "test-channel-3"

	// Set up expectations for the mock
	// When there's an error fetching the feed, we don't expect any store calls

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)
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

	// Set up expectations for the mock
	var capturedTimestamp time.Time
	mockStore.EXPECT().GetLastCheckedTimestamp(channelID).Return(time.Time{}, nil)
	mockStore.EXPECT().SetLastCheckedTimestamp(channelID, gomock.Any()).Do(func(channelID string, timestamp time.Time) {
		capturedTimestamp = timestamp
	})

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

	// Verify timestamp was set to the newest video time
	if !capturedTimestamp.Equal(newestVideoTime) {
		t.Errorf("Expected timestamp to be set to %v, but got %v", newestVideoTime, capturedTimestamp)
	}
}
