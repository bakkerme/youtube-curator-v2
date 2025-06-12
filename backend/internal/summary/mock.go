package summary

import (
	"context"
	"fmt"
	"strings"
	"time"

	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
)

// MockService provides a mock implementation for development and testing
type MockService struct {
	store store.Store
}

// NewMockService creates a new mock summary service
func NewMockService(store store.Store) *MockService {
	return &MockService{
		store: store,
	}
}

// GetOrGenerateSummary generates a mock summary for any video ID
func (ms *MockService) GetOrGenerateSummary(ctx context.Context, videoID string) *SummaryResult {
	result := &SummaryResult{
		VideoID: videoID,
	}

	// Check if we have existing summary in tracked videos
	summary, tracked := ms.findExistingSummary(videoID)
	if summary != nil {
		result.Summary = summary.Text
		result.SourceLanguage = summary.SourceLanguage
		result.GeneratedAt = summary.SummaryGeneratedAt
		result.Tracked = tracked
		return result
	}

	// Generate mock summary
	mockSummary := ms.generateMockSummary(videoID)

	result.Summary = mockSummary
	result.SourceLanguage = "en"
	result.GeneratedAt = time.Now()
	result.Tracked = false

	return result
}

// findExistingSummary looks for an existing summary in tracked videos
func (ms *MockService) findExistingSummary(videoID string) (*rss.Summary, bool) {
	// This is a placeholder - in a real implementation, you would
	// search through your video store or database to find existing summaries
	return nil, false
}

// generateMockSummary creates a mock summary based on video ID
func (ms *MockService) generateMockSummary(videoID string) string {
	// Safely get first 8 characters or the whole string if shorter
	shortID := videoID
	if len(videoID) > 8 {
		shortID = videoID[:8]
	}

	// Create different mock summaries based on video ID characteristics
	if strings.Contains(videoID, "tech") || strings.Contains(videoID, "code") {
		return fmt.Sprintf("This video explores advanced technical concepts and programming techniques. The presenter demonstrates practical examples and best practices for software development. Key insights include optimization strategies and modern development workflows. Video ID: %s", shortID)
	}

	if strings.Contains(videoID, "tutorial") || strings.Contains(videoID, "learn") {
		return fmt.Sprintf("An educational tutorial covering step-by-step instructions for beginners. The video breaks down complex topics into digestible segments with clear explanations. Viewers will gain practical skills and foundational knowledge. Video ID: %s", shortID)
	}

	if strings.Contains(videoID, "review") || strings.Contains(videoID, "test") {
		return fmt.Sprintf("A comprehensive review examining features, performance, and value proposition. The analysis includes detailed comparisons and real-world usage scenarios. The presenter provides honest insights and recommendations for potential users. Video ID: %s", shortID)
	}

	// Default mock summary
	return fmt.Sprintf("This video presents interesting content with valuable insights and engaging presentation. The creator shares expertise on the topic with clear explanations and practical examples. Viewers will find useful information and actionable takeaways. Mock summary for video ID: %s", shortID)
}
