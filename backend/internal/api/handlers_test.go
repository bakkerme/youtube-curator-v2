package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

// MockFeedProvider for testing
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
		return &rss.Feed{Entries: []rss.Entry{}}, nil // Return empty feed if not found
	}
	return feed, nil
}

// MockEmailSender for testing
type MockEmailSender struct{}

func (m *MockEmailSender) Send(to, subject, body string) error {
	return nil
}

// MockChannelProcessor for testing
type MockChannelProcessor struct{}

func (m *MockChannelProcessor) ProcessChannel(ctx context.Context, channelID string) processor.ChannelResult {
	return processor.ChannelResult{ChannelID: channelID}
}

func (m *MockChannelProcessor) ProcessChannelWithOptions(ctx context.Context, channelID string, ignoreLastChecked bool, maxItems int) processor.ChannelResult {
	return processor.ChannelResult{ChannelID: channelID}
}

func TestGetVideos_EmptyCache_FetchesFromChannels(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	mockEmailSender := &MockEmailSender{}
	mockProcessor := &MockChannelProcessor{}
	cfg := &config.Config{}

	// Create video store with short TTL for testing
	videoStore := store.NewVideoStore(1 * time.Hour)

	// Setup mock expectations
	channels := []store.Channel{
		{ID: "channel1", Title: "Test Channel 1"},
		{ID: "channel2", Title: "Test Channel 2"},
	}
	mockStore.EXPECT().GetChannels().Return(channels, nil).Times(1)

	// Setup mock feeds
	video1 := rss.Entry{
		ID:        "video1",
		Title:     "Test Video 1",
		Published: time.Now().Add(-1 * time.Hour),
	}
	video2 := rss.Entry{
		ID:        "video2",
		Title:     "Test Video 2",
		Published: time.Now().Add(-2 * time.Hour),
	}

	mockFeedProvider.feeds["channel1"] = &rss.Feed{
		Entries: []rss.Entry{video1},
	}
	mockFeedProvider.feeds["channel2"] = &rss.Feed{
		Entries: []rss.Entry{video2},
	}

	// Create handlers
	handlers := NewHandlers(mockStore, mockFeedProvider, mockEmailSender, cfg, mockProcessor, videoStore)

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/videos", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handlers.GetVideos(c)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	// Verify videos were fetched and stored
	videos := videoStore.GetAllVideos()
	if len(videos) != 2 {
		t.Fatalf("Expected 2 videos in store, got %d", len(videos))
	}
}

func TestGetVideos_WithRefreshParameter(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	mockEmailSender := &MockEmailSender{}
	mockProcessor := &MockChannelProcessor{}
	cfg := &config.Config{}

	// Create video store and pre-populate it
	videoStore := store.NewVideoStore(1 * time.Hour)
	oldVideo := rss.Entry{
		ID:        "old_video",
		Title:     "Old Video",
		Published: time.Now().Add(-3 * time.Hour),
	}
	videoStore.AddVideo("channel1", oldVideo)

	// Setup mock expectations - should be called when refresh=true
	channels := []store.Channel{
		{ID: "channel1", Title: "Test Channel 1"},
	}
	mockStore.EXPECT().GetChannels().Return(channels, nil).Times(1)

	// Setup new video in mock feed
	newVideo := rss.Entry{
		ID:        "new_video",
		Title:     "New Video",
		Published: time.Now().Add(-1 * time.Hour),
	}
	mockFeedProvider.feeds["channel1"] = &rss.Feed{
		Entries: []rss.Entry{newVideo},
	}

	// Create handlers
	handlers := NewHandlers(mockStore, mockFeedProvider, mockEmailSender, cfg, mockProcessor, videoStore)

	// Create test request with refresh=true
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/videos?refresh=true", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handlers.GetVideos(c)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	// Verify new video was fetched and stored
	videos := videoStore.GetAllVideos()
	if len(videos) != 2 {
		t.Fatalf("Expected 2 videos in store, got %d", len(videos))
	}

	// Check that we have both old and new videos
	foundOld := false
	foundNew := false
	for _, video := range videos {
		if video.Entry.ID == "old_video" {
			foundOld = true
		}
		if video.Entry.ID == "new_video" {
			foundNew = true
		}
	}

	if !foundOld {
		t.Error("Expected to find old video in store")
	}
	if !foundNew {
		t.Error("Expected to find new video in store")
	}
}

func TestGetVideos_UsesCache_WhenNotExpired(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockStore(ctrl)
	mockFeedProvider := NewMockFeedProvider()
	mockEmailSender := &MockEmailSender{}
	mockProcessor := &MockChannelProcessor{}
	cfg := &config.Config{}

	// Create video store and pre-populate it
	videoStore := store.NewVideoStore(1 * time.Hour)
	cachedVideo := rss.Entry{
		ID:        "cached_video",
		Title:     "Cached Video",
		Published: time.Now().Add(-30 * time.Minute), // Not expired
	}
	videoStore.AddVideo("channel1", cachedVideo)

	// Create handlers
	handlers := NewHandlers(mockStore, mockFeedProvider, mockEmailSender, cfg, mockProcessor, videoStore)

	// Create test request without refresh parameter
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/videos", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handlers.GetVideos(c)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	// Verify cached video is returned
	videos := videoStore.GetAllVideos()
	if len(videos) != 1 {
		t.Fatalf("Expected 1 video from cache, got %d", len(videos))
	}

	if videos[0].Entry.ID != "cached_video" {
		t.Errorf("Expected cached video ID 'cached_video', got '%s'", videos[0].Entry.ID)
	}
}
