package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
	"youtube-curator-v2/internal/ytdlp"

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
	videoStore := store.NewVideoStore(1 * time.Hour)
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider, videoStore)

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
	videoStore := store.NewVideoStore(1 * time.Hour)
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider, videoStore)

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
	videoStore := store.NewVideoStore(1 * time.Hour)
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider, videoStore)

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
	videoStore := store.NewVideoStore(1 * time.Hour)
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider, videoStore)

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

func testProcessChannelWithEnrichment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	mockVideoStore := store.NewVideoStore(1 * time.Hour)

	// Create processor with mock enricher
	processor := &DefaultChannelProcessor{
		db:           mockStore,
		feedProvider: mockFeedProvider,
		videoStore:   mockVideoStore,
		enricher:     ytdlp.NewMockEnricher(),
	}

	// Set up mock data
	channelID := "UC123456789012345678901"
	testTime := time.Now().Add(-1 * time.Hour)

	// Mock feed with one entry
	mockFeed := &rss.Feed{
		Title: "Test Channel",
		Entries: []rss.Entry{
			{
				Title:     "Test Video",
				ID:        "yt:video:dQw4w9WgXcQ",
				Published: time.Now(),
				Link:      rss.Link{Href: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"},
				Content:   "Test video content",
				Author:    rss.Author{Name: "Test Author"},
			},
		},
	}

	// Set up mock expectations
	mockFeedProvider.feeds[channelID] = mockFeed
	mockStore.EXPECT().GetLastCheckedTimestamp(channelID).Return(testTime, nil)
	mockStore.EXPECT().SetLastCheckedTimestamp(channelID, gomock.Any())

	// Process the channel
	ctx := context.Background()
	result := processor.ProcessChannel(ctx, channelID)

	// Verify results
	if result.Error != nil {
		t.Fatalf("ProcessChannel returned error: %v", result.Error)
	}

	if result.NewVideo == nil {
		t.Fatal("Expected new video but got nil")
	}

	// Verify enrichment worked
	if result.NewVideo.Duration == 0 {
		t.Error("Expected duration to be set by enricher")
	}

	if len(result.NewVideo.Tags) == 0 {
		t.Error("Expected tags to be set by enricher")
	}

	if len(result.NewVideo.TopComments) == 0 {
		t.Error("Expected top comments to be set by enricher")
	}

	if result.NewVideo.AutoSubtitles == "" {
		t.Error("Expected auto subtitles to be set by enricher")
	}

	// Verify specific mock values
	expectedDuration := 300
	if result.NewVideo.Duration != expectedDuration {
		t.Errorf("Expected duration %d, got %d", expectedDuration, result.NewVideo.Duration)
	}

	expectedTags := []string{"technology", "tutorial", "programming", "demo"}
	if len(result.NewVideo.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(result.NewVideo.Tags))
	}
}

func TestProcessChannelWithOptions_IgnoreLastChecked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	videoStore := store.NewVideoStore(1 * time.Hour)
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider, videoStore)

	channelID := "test-channel-ignore"
	lastChecked := time.Now().Add(-24 * time.Hour)

	// Create a feed with videos older than last checked
	oldVideoTime := lastChecked.Add(-48 * time.Hour)
	mockFeedProvider.feeds[channelID] = &rss.Feed{
		Entries: []rss.Entry{
			{
				Title:     "Old Video",
				Published: oldVideoTime,
				ID:        "old-video-id",
			},
		},
	}

	// Set up expectations for the mock
	mockStore.EXPECT().GetLastCheckedTimestamp(channelID).Return(lastChecked, nil)
	// SetLastCheckedTimestamp should NOT be called when ignoreLastChecked is true

	// Execute with ignoreLastChecked = true
	result := processor.ProcessChannelWithOptions(context.Background(), channelID, true, 0)

	// Verify
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	// With ignoreLastChecked=true, even old videos should be considered "new"
	if result.NewVideo == nil {
		t.Fatal("Expected a new video when ignoring last checked, but got nil")
	}
	if result.NewVideo.Title != "Old Video" {
		t.Errorf("Expected video title 'Old Video', but got '%s'", result.NewVideo.Title)
	}
	if result.ChannelID != channelID {
		t.Errorf("Expected channel ID %s, but got %s", channelID, result.ChannelID)
	}
}

func TestProcessChannelWithOptions_MaxItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	videoStore := store.NewVideoStore(1 * time.Hour)
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider, videoStore)

	channelID := "test-channel-maxitems"
	lastChecked := time.Now().Add(-24 * time.Hour)

	// Create a feed with multiple new videos
	mockFeedProvider.feeds[channelID] = &rss.Feed{
		Entries: []rss.Entry{
			{
				Title:     "Video 1",
				Published: time.Now().Add(-1 * time.Hour),
				ID:        "video-1",
			},
			{
				Title:     "Video 2",
				Published: time.Now().Add(-2 * time.Hour),
				ID:        "video-2",
			},
			{
				Title:     "Video 3",
				Published: time.Now().Add(-3 * time.Hour),
				ID:        "video-3",
			},
			{
				Title:     "Video 4",
				Published: time.Now().Add(-4 * time.Hour),
				ID:        "video-4",
			},
		},
	}

	// Set up expectations for the mock
	mockStore.EXPECT().GetLastCheckedTimestamp(channelID).Return(lastChecked, nil)
	mockStore.EXPECT().SetLastCheckedTimestamp(channelID, gomock.Any())

	// Execute with maxItems = 2
	result := processor.ProcessChannelWithOptions(context.Background(), channelID, false, 2)

	// Verify
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	if result.NewVideo == nil {
		t.Fatal("Expected a new video, but got nil")
	}
	// Should return the latest video (Video 1) since we process in order and stop at maxItems
	if result.NewVideo.Title != "Video 1" {
		t.Errorf("Expected newest video title 'Video 1', but got '%s'", result.NewVideo.Title)
	}
	if result.ChannelID != channelID {
		t.Errorf("Expected channel ID %s, but got %s", channelID, result.ChannelID)
	}
}

func TestProcessChannelWithOptions_MaxItemsZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	videoStore := store.NewVideoStore(1 * time.Hour)
	processor := NewDefaultChannelProcessor(mockStore, mockFeedProvider, videoStore)

	channelID := "test-channel-maxitems-zero"
	lastChecked := time.Now().Add(-24 * time.Hour)

	// Create a feed with multiple new videos
	mockFeedProvider.feeds[channelID] = &rss.Feed{
		Entries: []rss.Entry{
			{
				Title:     "Video 1",
				Published: time.Now().Add(-1 * time.Hour),
				ID:        "video-1",
			},
			{
				Title:     "Video 2",
				Published: time.Now().Add(-2 * time.Hour),
				ID:        "video-2",
			},
			{
				Title:     "Video 3",
				Published: time.Now().Add(-3 * time.Hour),
				ID:        "video-3",
			},
		},
	}

	// Set up expectations for the mock
	var capturedTimestamp time.Time
	mockStore.EXPECT().GetLastCheckedTimestamp(channelID).Return(lastChecked, nil)
	mockStore.EXPECT().SetLastCheckedTimestamp(channelID, gomock.Any()).Do(func(channelID string, timestamp time.Time) {
		capturedTimestamp = timestamp
	})

	// Execute with maxItems = 0 (should process all videos)
	result := processor.ProcessChannelWithOptions(context.Background(), channelID, false, 0)

	// Verify
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	if result.NewVideo == nil {
		t.Fatal("Expected a new video, but got nil")
	}
	// Should return the latest video (Video 1) and process all videos
	if result.NewVideo.Title != "Video 1" {
		t.Errorf("Expected newest video title 'Video 1', but got '%s'", result.NewVideo.Title)
	}

	// Verify timestamp was updated (meaning all videos were processed)
	expectedTime := time.Now().Add(-1 * time.Hour)
	if capturedTimestamp.Sub(expectedTime) > time.Minute {
		t.Errorf("Expected timestamp to be close to %v, but got %v", expectedTime, capturedTimestamp)
	}
}
