package main

import (
	"context"
	"testing"
	"time"

	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
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
		Channels:       []string{"channel-1", "channel-2", "channel-3"},
		RecipientEmail: "test@example.com",
	}

	mockEmailSender := NewMockEmailSender()
	mockProcessor := NewMockChannelProcessor()

	// All channels return no new videos
	for _, channelID := range cfg.Channels {
		mockProcessor.results[channelID] = processor.ChannelResult{
			ChannelID: channelID,
			NewVideo:  nil,
			Error:     nil,
		}
	}

	// Execute
	checkForNewVideos(cfg, mockEmailSender, mockProcessor)

	// Verify - no emails should be sent
	if len(mockEmailSender.sentEmails) != 0 {
		t.Errorf("Expected no emails to be sent, but got %d", len(mockEmailSender.sentEmails))
	}
}

func TestCheckForNewVideos_WithNewVideos(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Channels:       []string{"channel-1", "channel-2", "channel-3"},
		RecipientEmail: "test@example.com",
	}

	mockEmailSender := NewMockEmailSender()
	mockProcessor := NewMockChannelProcessor()

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
	checkForNewVideos(cfg, mockEmailSender, mockProcessor)

	// Verify
	if len(mockEmailSender.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, but got %d", len(mockEmailSender.sentEmails))
	}

	sentEmail := mockEmailSender.sentEmails[0]
	if sentEmail.Recipient != cfg.RecipientEmail {
		t.Errorf("Expected email recipient to be %s, but got %s", cfg.RecipientEmail, sentEmail.Recipient)
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
		Channels:       []string{"channel-1", "channel-2", "channel-3"},
		RecipientEmail: "test@example.com",
	}

	mockEmailSender := NewMockEmailSender()
	mockProcessor := NewMockChannelProcessor()

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
	checkForNewVideos(cfg, mockEmailSender, mockProcessor)

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
