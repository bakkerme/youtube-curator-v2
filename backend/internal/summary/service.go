package summary

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"youtube-curator-v2/internal/customerrors"
	"youtube-curator-v2/internal/openai"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
	"youtube-curator-v2/internal/ytdlp"
)

// SummaryServiceInterface defines the interface that both real and mock services implement
type SummaryServiceInterface interface {
	GetOrGenerateSummary(ctx context.Context, videoID string) *SummaryResult
}

// SummaryResult represents the result of a summarization operation
type SummaryResult struct {
	VideoID        string
	Summary        string
	SourceLanguage string
	GeneratedAt    time.Time
	Tracked        bool
	Error          error
}

// Service handles video summarization operations
type Service struct {
	store         store.Store
	ytdlpEnricher ytdlp.Enricher
	openaiClient  openai.OpenAIClient
}

// NewService creates a new summary service
func NewService(store store.Store, ytdlpEnricher ytdlp.Enricher, openaiClient openai.OpenAIClient) *Service {
	return &Service{
		store:         store,
		ytdlpEnricher: ytdlpEnricher,
		openaiClient:  openaiClient,
	}
}

// GetOrGenerateSummary retrieves an existing summary or generates a new one
func (s *Service) GetOrGenerateSummary(ctx context.Context, videoID string) *SummaryResult {
	result := &SummaryResult{
		VideoID: videoID,
	}

	// Check if we have LLM configuration
	llmConfig, err := s.store.GetLLMConfig()
	if err != nil {
		result.Error = fmt.Errorf("failed to get LLM configuration: %w", err)
		return result
	}
	if llmConfig == nil || llmConfig.EndpointURL == "" {
		result.Error = fmt.Errorf("LLM not configured")
		return result
	}

	// Try to find existing summary in tracked videos first
	// This is a simplified approach - in a real implementation,
	// you might want to search through stored videos more efficiently
	summary, tracked := s.findExistingSummary(videoID)
	if summary != nil {
		result.Summary = summary.Text
		result.SourceLanguage = summary.SourceLanguage
		result.GeneratedAt = summary.SummaryGeneratedAt
		result.Tracked = tracked
		return result
	}

	// Generate new summary
	generatedSummary, sourceLanguage, err := s.generateSummary(ctx, videoID, llmConfig)
	if err != nil {
		result.Error = err
		return result
	}

	result.Summary = generatedSummary
	result.SourceLanguage = sourceLanguage
	result.GeneratedAt = time.Now()
	result.Tracked = false // For now, arbitrary videos are not tracked

	return result
}

// findExistingSummary looks for an existing summary in tracked videos
func (s *Service) findExistingSummary(videoID string) (*rss.Summary, bool) {
	// This is a placeholder - in a real implementation, you would
	// search through your video store or database to find existing summaries
	// For now, we'll return nil to always generate new summaries
	return nil, false
}

// generateSummary generates a new summary for a video
func (s *Service) generateSummary(ctx context.Context, videoID string, llmConfig *store.LLMConfig) (string, string, error) {
	// Create a temporary entry to enrich with subtitles
	entry := &rss.Entry{
		ID: fmt.Sprintf("yt:video:%s", videoID),
	}

	// Use yt-dlp to fetch subtitles
	err := s.ytdlpEnricher.EnrichEntry(ctx, entry)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	// Check if we have subtitles
	if entry.AutoSubtitles == "" {
		return "", "", fmt.Errorf("no subtitles available for video")
	}

	// Fetch and parse actual subtitle content from the URL
	subtitleText, err := s.fetchSubtitleText(ctx, entry.AutoSubtitles)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch subtitle content: %w", err)
	}

	// Generate summary using LLM
	summary, err := s.callLLMForSummary(ctx, subtitleText, llmConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate summary: %w", err)
	}

	// Assume English for now - in a real implementation, you'd detect language
	sourceLanguage := "en"

	return summary, sourceLanguage, nil
}

// callLLMForSummary calls the LLM to generate a summary
func (s *Service) callLLMForSummary(ctx context.Context, subtitleText string, llmConfig *store.LLMConfig) (string, error) {
	// Create OpenAI client with the configured settings
	client := openai.New(llmConfig.EndpointURL, llmConfig.APIKey, llmConfig.Model)

	systemPrompt := `You are a helpful assistant that creates concise, informative summaries of YouTube video content based on their subtitles. 

Please create a summary that:
- Captures the main points and key insights
- Is approximately 2-4 sentences long
- Focuses on the most important information
- Is written in a clear, engaging style`

	userPrompt := fmt.Sprintf("Please summarize this YouTube video based on its subtitles:\n\n%s", subtitleText)

	// Create a channel to receive the result
	resultChan := make(chan customerrors.ErrorString, 1)

	// Call the LLM
	go client.ChatCompletion(
		ctx,
		systemPrompt,
		[]string{userPrompt},
		nil, // no images
		nil, // no schema
		0.7, // temperature
		0,   // no max tokens limit
		resultChan,
	)

	// Wait for the result
	select {
	case result := <-resultChan:
		if result.Err != nil {
			return "", result.Err
		}
		return result.Value, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// fetchSubtitleText fetches and parses subtitle content from a URL
func (s *Service) fetchSubtitleText(ctx context.Context, subtitleURL string) (string, error) {
	if subtitleURL == "" {
		return "", fmt.Errorf("subtitle URL is empty")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", subtitleURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set user agent to avoid potential blocking
	req.Header.Set("User-Agent", "youtube-curator/1.0")

	// Fetch subtitle content
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subtitle content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse subtitle content based on format
	subtitleText := s.parseSubtitleContent(string(body))

	if subtitleText == "" {
		return "", fmt.Errorf("no text content found in subtitles")
	}

	return subtitleText, nil
}

// parseSubtitleContent parses subtitle content from VTT or SRT format
func (s *Service) parseSubtitleContent(content string) string {
	// Remove VTT header and timing information
	// VTT format typically starts with "WEBVTT" and has timing lines like "00:00:01.000 --> 00:00:04.000"
	// SRT format has numbered blocks with timing lines like "00:00:01,000 --> 00:00:04,000"

	lines := strings.Split(content, "\n")
	var textLines []string

	// Regex patterns for timing lines
	vttTimingPattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} --> \d{2}:\d{2}:\d{2}\.\d{3}`)
	srtTimingPattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2},\d{3} --> \d{2}:\d{2}:\d{2},\d{3}`)

	numericPattern := regexp.MustCompile(`^\d+$`) // Matches numeric lines (SRT sequence numbers)
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip VTT header
		if strings.HasPrefix(line, "WEBVTT") {
			continue
		}

		// Skip timing lines
		if vttTimingPattern.MatchString(line) || srtTimingPattern.MatchString(line) {
			continue
		}

		// Skip numeric lines (SRT sequence numbers)
		if numericPattern.MatchString(line) {
			continue
		}

		// Skip VTT style/position information
		if strings.Contains(line, "align:") || strings.Contains(line, "position:") || strings.Contains(line, "size:") {
			continue
		}

		// Clean up HTML tags and formatting that might be in subtitles
		line = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(line, "")
		line = strings.ReplaceAll(line, "&amp;", "&")
		line = strings.ReplaceAll(line, "&lt;", "<")
		line = strings.ReplaceAll(line, "&gt;", ">")
		line = strings.ReplaceAll(line, "&quot;", "\"")
		line = strings.ReplaceAll(line, "&#39;", "'")
		line = strings.ReplaceAll(line, "&apos;", "'")

		// Add to text lines if it contains actual content
		if line != "" {
			textLines = append(textLines, line)
		}
	}

	// Join all text lines with spaces
	return strings.Join(textLines, " ")
}
