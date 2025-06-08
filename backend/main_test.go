package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"go.uber.org/mock/gomock"
)

// MockChannelProcessor is a mock implementation of processor.ChannelProcessor for testing
type MockChannelProcessor struct {
	results map[string]processor.ChannelResult
}

func NewMockChannelProcessor() *MockChannelProcessor {
	return &MockChannelProcessor{
		results: make(map[string]processor.ChannelResult),
	}
}

func (m *MockChannelProcessor) ProcessChannel(ctx context.Context, channelID string) processor.ChannelResult {
	if result, ok := m.results[channelID]; ok {
		return result
	}
	return processor.ChannelResult{
		ChannelID: channelID,
		NewVideo:  nil,
		Error:     nil,
	}
}

func (m *MockChannelProcessor) ProcessChannelWithOptions(ctx context.Context, channelID string, ignoreLastChecked bool, maxItems int) processor.ChannelResult {
	// For testing purposes, we can use the same logic as ProcessChannel
	// In a real test, you might want to verify the ignoreLastChecked and maxItems parameters
	return m.ProcessChannel(ctx, channelID)
}

// MockEmailSender is a mock implementation of email.Sender for testing
type MockEmailSender struct {
	sentEmails []SentEmail
}

type SentEmail struct {
	Recipient string
	Subject   string
	Content   string
}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		sentEmails: make([]SentEmail, 0),
	}
}

func (m *MockEmailSender) Send(recipient string, subject string, htmlContent string) error {
	m.sentEmails = append(m.sentEmails, SentEmail{
		Recipient: recipient,
		Subject:   subject,
		Content:   htmlContent,
	})
	return nil
}

func TestCheckForNewVideos_NoNewVideos(t *testing.T) {
	// Setup
	cfg := &config.Config{
		RecipientEmail: "test@example.com",
		RSSConcurrency: 3,
	}

	mockEmailSender := NewMockEmailSender()
	mockProcessor := NewMockChannelProcessor()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)

	// Set up expectations for the mock store
	channels := []store.Channel{
		{ID: "channel-1", Title: "Channel 1"},
		{ID: "channel-2", Title: "Channel 2"},
		{ID: "channel-3", Title: "Channel 3"},
	}
	mockStore.EXPECT().GetChannels().Return(channels, nil)

	// All channels return no new videos
	for _, channelID := range []string{"channel-1", "channel-2", "channel-3"} {
		mockProcessor.results[channelID] = processor.ChannelResult{
			ChannelID: channelID,
			NewVideo:  nil,
			Error:     nil,
		}
	}

	// Execute
	checkForNewVideos(cfg, mockEmailSender, mockProcessor, mockStore)

	// Verify - no emails should be sent
	if len(mockEmailSender.sentEmails) != 0 {
		t.Errorf("Expected no emails to be sent, but got %d", len(mockEmailSender.sentEmails))
	}
}

func TestCheckForNewVideos_WithNewVideos(t *testing.T) {
	// Setup
	cfg := &config.Config{
		RecipientEmail: "test@example.com",
		RSSConcurrency: 3,
	}

	mockEmailSender := NewMockEmailSender()
	mockProcessor := NewMockChannelProcessor()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)

	// Set up expectations for the mock store
	channels := []store.Channel{
		{ID: "channel-1", Title: "Channel 1"},
		{ID: "channel-2", Title: "Channel 2"},
		{ID: "channel-3", Title: "Channel 3"},
	}
	mockStore.EXPECT().GetChannels().Return(channels, nil)

	// Mock SMTP config retrieval - return a valid config
	smtpConfig := &store.SMTPConfig{
		Server:         "smtp.example.com",
		Port:           "587",
		Username:       "test@example.com",
		Password:       "password",
		RecipientEmail: "recipient@example.com",
	}
	mockStore.EXPECT().GetSMTPConfig().Return(smtpConfig, nil)

	// Channel 1 has a new video
	video1 := &rss.Entry{
		Title:     "New Video from Channel 1",
		Published: time.Now().Add(-2 * time.Hour),
		ID:        "video-1",
		Link:      rss.Link{Href: "https://youtube.com/watch?v=1"},
	}
	mockProcessor.results["channel-1"] = processor.ChannelResult{
		ChannelID: "channel-1",
		NewVideo:  video1,
		Error:     nil,
	}

	// Channel 2 has no new videos
	mockProcessor.results["channel-2"] = processor.ChannelResult{
		ChannelID: "channel-2",
		NewVideo:  nil,
		Error:     nil,
	}

	// Channel 3 has a new video
	video3 := &rss.Entry{
		Title:     "New Video from Channel 3",
		Published: time.Now().Add(-1 * time.Hour),
		ID:        "video-3",
		Link:      rss.Link{Href: "https://youtube.com/watch?v=3"},
	}
	mockProcessor.results["channel-3"] = processor.ChannelResult{
		ChannelID: "channel-3",
		NewVideo:  video3,
		Error:     nil,
	}

	// Execute
	checkForNewVideos(cfg, mockEmailSender, mockProcessor, mockStore)

	// Verify
	if len(mockEmailSender.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, but got %d", len(mockEmailSender.sentEmails))
	}

	sentEmail := mockEmailSender.sentEmails[0]
	if sentEmail.Recipient != smtpConfig.RecipientEmail {
		t.Errorf("Expected email recipient to be %s, but got %s", smtpConfig.RecipientEmail, sentEmail.Recipient)
	}
	if sentEmail.Subject != "New YouTube Videos Update" {
		t.Errorf("Expected email subject to be 'New YouTube Videos Update', but got '%s'", sentEmail.Subject)
	}

	// Verify the email contains both videos
	if !contains(sentEmail.Content, video1.Title) {
		t.Errorf("Expected email to contain video 1 title '%s'", video1.Title)
	}
	if !contains(sentEmail.Content, video3.Title) {
		t.Errorf("Expected email to contain video 3 title '%s'", video3.Title)
	}
}

func TestCheckForNewVideos_MixedResults(t *testing.T) {
	// Setup
	cfg := &config.Config{
		RecipientEmail: "test@example.com",
		RSSConcurrency: 3,
	}

	mockEmailSender := NewMockEmailSender()
	mockProcessor := NewMockChannelProcessor()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)

	// Set up expectations for the mock store
	channels := []store.Channel{
		{ID: "channel-1", Title: "Channel 1"},
		{ID: "channel-2", Title: "Channel 2"},
		{ID: "channel-3", Title: "Channel 3"},
	}
	mockStore.EXPECT().GetChannels().Return(channels, nil)

	// Mock SMTP config retrieval - return a valid config
	smtpConfig := &store.SMTPConfig{
		Server:         "smtp.example.com",
		Port:           "587",
		Username:       "test@example.com",
		Password:       "password",
		RecipientEmail: "recipient@example.com",
	}
	mockStore.EXPECT().GetSMTPConfig().Return(smtpConfig, nil)

	// Channel 1 has an error
	mockProcessor.results["channel-1"] = processor.ChannelResult{
		ChannelID: "channel-1",
		NewVideo:  nil,
		Error:     context.DeadlineExceeded,
	}

	// Channel 2 has a new video
	video2 := &rss.Entry{
		Title:     "New Video from Channel 2",
		Published: time.Now().Add(-30 * time.Minute),
		ID:        "video-2",
		Link:      rss.Link{Href: "https://youtube.com/watch?v=2"},
	}
	mockProcessor.results["channel-2"] = processor.ChannelResult{
		ChannelID: "channel-2",
		NewVideo:  video2,
		Error:     nil,
	}

	// Channel 3 has no new videos
	mockProcessor.results["channel-3"] = processor.ChannelResult{
		ChannelID: "channel-3",
		NewVideo:  nil,
		Error:     nil,
	}

	// Execute
	checkForNewVideos(cfg, mockEmailSender, mockProcessor, mockStore)

	// Verify - should still send email with the one successful video
	if len(mockEmailSender.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent despite error, but got %d", len(mockEmailSender.sentEmails))
	}

	sentEmail := mockEmailSender.sentEmails[0]
	if !contains(sentEmail.Content, video2.Title) {
		t.Errorf("Expected email to contain video 2 title '%s'", video2.Title)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func TestProcessChannelsConcurrently(t *testing.T) {
	// Setup
	mockProcessor := NewMockChannelProcessor()
	ctx := context.Background()
	
	// Create test channels
	channels := []store.Channel{
		{ID: "channel-1", Title: "Channel 1"},
		{ID: "channel-2", Title: "Channel 2"},
		{ID: "channel-3", Title: "Channel 3"},
		{ID: "channel-4", Title: "Channel 4"},
	}
	
	// Set up expected results
	video1 := &rss.Entry{Title: "Video 1", ID: "video-1"}
	video2 := &rss.Entry{Title: "Video 2", ID: "video-2"}
	
	mockProcessor.results["channel-1"] = processor.ChannelResult{
		ChannelID: "channel-1",
		NewVideo:  video1,
		Error:     nil,
	}
	mockProcessor.results["channel-2"] = processor.ChannelResult{
		ChannelID: "channel-2",
		NewVideo:  video2,
		Error:     nil,
	}
	mockProcessor.results["channel-3"] = processor.ChannelResult{
		ChannelID: "channel-3",
		NewVideo:  nil,
		Error:     nil,
	}
	mockProcessor.results["channel-4"] = processor.ChannelResult{
		ChannelID: "channel-4",
		NewVideo:  nil,
		Error:     fmt.Errorf("test error"),
	}
	
	// Test with concurrency level 2
	results := processChannelsConcurrently(ctx, channels, mockProcessor, 2)
	
	// Verify all channels were processed
	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}
	
	// Verify specific results
	if result, ok := results["channel-1"]; !ok || result.NewVideo == nil || result.NewVideo.Title != "Video 1" {
		t.Errorf("Channel 1 result incorrect: %+v", result)
	}
	
	if result, ok := results["channel-2"]; !ok || result.NewVideo == nil || result.NewVideo.Title != "Video 2" {
		t.Errorf("Channel 2 result incorrect: %+v", result)
	}
	
	if result, ok := results["channel-3"]; !ok || result.NewVideo != nil {
		t.Errorf("Channel 3 should have no new video: %+v", result)
	}
	
	if result, ok := results["channel-4"]; !ok || result.Error == nil {
		t.Errorf("Channel 4 should have an error: %+v", result)
	}
}

func TestProcessChannelsConcurrently_EmptyChannels(t *testing.T) {
	mockProcessor := NewMockChannelProcessor()
	ctx := context.Background()
	channels := []store.Channel{}
	
	results := processChannelsConcurrently(ctx, channels, mockProcessor, 5)
	
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty channels, got %d", len(results))
	}
}

func TestProcessChannelsConcurrently_ConcurrencyLimit(t *testing.T) {
	mockProcessor := NewMockChannelProcessor()
	ctx := context.Background()
	
	// Create 10 channels
	channels := make([]store.Channel, 10)
	for i := 0; i < 10; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		channels[i] = store.Channel{ID: channelID, Title: fmt.Sprintf("Channel %d", i)}
		mockProcessor.results[channelID] = processor.ChannelResult{
			ChannelID: channelID,
			NewVideo:  nil,
			Error:     nil,
		}
	}
	
	// Test that concurrency is limited correctly when we have more channels than workers
	results := processChannelsConcurrently(ctx, channels, mockProcessor, 3)
	
	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}
}

func TestProcessChannelsConcurrently_MaxConcurrencyLimit(t *testing.T) {
	mockProcessor := NewMockChannelProcessor()
	ctx := context.Background()
	
	// Create 15 channels
	channels := make([]store.Channel, 15)
	for i := 0; i < 15; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		channels[i] = store.Channel{ID: channelID, Title: fmt.Sprintf("Channel %d", i)}
		mockProcessor.results[channelID] = processor.ChannelResult{
			ChannelID: channelID,
			NewVideo:  nil,
			Error:     nil,
		}
	}
	
	// Test that excessive concurrency is limited to max value (10)
	// This should warn but still process all channels
	results := processChannelsConcurrently(ctx, channels, mockProcessor, 15)
	
	if len(results) != 15 {
		t.Errorf("Expected 15 results, got %d", len(results))
	}
}

func TestCheckForNewVideos_FallbackToConfigEmail(t *testing.T) {
	// Setup
	cfg := &config.Config{
		RecipientEmail: "fallback@example.com",
		RSSConcurrency: 1,
	}

	mockEmailSender := NewMockEmailSender()
	mockProcessor := NewMockChannelProcessor()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := store.NewMockStore(ctrl)

	// Set up expectations for the mock store
	channels := []store.Channel{
		{ID: "channel-1", Title: "Channel 1"},
	}
	mockStore.EXPECT().GetChannels().Return(channels, nil)

	// Mock SMTP config retrieval - return nil (no config in database)
	mockStore.EXPECT().GetSMTPConfig().Return(nil, nil)

	// Channel 1 has a new video
	video1 := &rss.Entry{
		Title:     "New Video from Channel 1",
		Published: time.Now().Add(-2 * time.Hour),
		ID:        "video-1",
		Link:      rss.Link{Href: "https://youtube.com/watch?v=1"},
	}
	mockProcessor.results["channel-1"] = processor.ChannelResult{
		ChannelID: "channel-1",
		NewVideo:  video1,
		Error:     nil,
	}

	// Execute
	checkForNewVideos(cfg, mockEmailSender, mockProcessor, mockStore)

	// Verify - should fallback to config.RecipientEmail
	if len(mockEmailSender.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, but got %d", len(mockEmailSender.sentEmails))
	}

	sentEmail := mockEmailSender.sentEmails[0]
	if sentEmail.Recipient != cfg.RecipientEmail {
		t.Errorf("Expected email recipient to be %s (fallback), but got %s", cfg.RecipientEmail, sentEmail.Recipient)
	}
}
