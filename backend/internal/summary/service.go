package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
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
	Thinking       string
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
	generatedSummary, thinking, sourceLanguage, err := s.generateSummary(ctx, videoID, llmConfig)
	if err != nil {
		result.Error = err
		return result
	}

	result.Summary = generatedSummary
	result.Thinking = thinking
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
func (s *Service) generateSummary(ctx context.Context, videoID string, llmConfig *store.LLMConfig) (string, string, string, error) {
	// Create a temporary entry to enrich with subtitles
	entry := &rss.Entry{
		ID: videoID, // videoID should already be in yt:video:ID format from the handler
	}

	// Use yt-dlp to fetch subtitles
	err := s.ytdlpEnricher.EnrichEntry(ctx, entry)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	// Check if we have subtitles
	if entry.AutoSubtitles == "" {
		return "", "", "", fmt.Errorf("no subtitles available for video")
	}

	// Fetch and parse actual subtitle content from the URL
	subtitleText, err := s.fetchSubtitleText(ctx, entry.AutoSubtitles)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch subtitle content: %w", err)
	}

	// Generate summary using LLM
	rawResponse, err := s.callLLMForSummary(ctx, subtitleText, llmConfig)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate summary: %w", err)
	}

	// Parse thinking blocks from the response
	thinking, summary := parseThinkingBlocks(rawResponse)

	// Assume English for now - in a real implementation, you'd detect language
	sourceLanguage := "en"

	return summary, thinking, sourceLanguage, nil
}

// callLLMForSummary calls the LLM to generate a summary
func (s *Service) callLLMForSummary(ctx context.Context, subtitleText string, llmConfig *store.LLMConfig) (string, error) {
	// Create OpenAI client with the configured settings
	client := openai.New(llmConfig.EndpointURL, llmConfig.APIKey, llmConfig.Model)

	systemPrompt := `Summarize the provided YouTube video, providing all the key points of the video and any related insights. Don't be afraid to go in-depth with the details.`

	userPrompt := subtitleText

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

	// Optimize text for token efficiency
	optimizedText := s.optimizeSubtitleText(subtitleText)

	return optimizedText, nil
}

// parseSubtitleContent parses subtitle content from VTT, SRT, or M3U8 format
func (s *Service) parseSubtitleContent(content string) string {
	// Check if this is an M3U8 playlist file
	if strings.HasPrefix(strings.TrimSpace(content), "#EXTM3U") {
		return s.parseM3U8Playlist(content)
	}

	// For VTT/SRT content, use the dedicated parser
	return s.parseVTTOrSRTContent(content)
}

// parseM3U8Playlist parses an M3U8 playlist and fetches subtitle segments
func (s *Service) parseM3U8Playlist(content string) string {
	lines := strings.Split(content, "\n")
	var segmentURLs []string

	// Extract segment URLs from M3U8 playlist
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip comments and metadata lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		
		// This should be a URL to a subtitle segment
		if strings.HasPrefix(line, "http") {
			segmentURLs = append(segmentURLs, line)
		}
	}

	// If no segments found, return empty string
	if len(segmentURLs) == 0 {
		return ""
	}

	// For now, fetch only the first segment to avoid complexity
	// In a full implementation, you might want to fetch all segments
	ctx := context.Background()
	
	segmentContent, err := s.fetchSegmentContent(ctx, segmentURLs[0])
	if err != nil {
		// If we can't fetch the segment, return empty string
		return ""
	}
	
	return segmentContent
}

// fetchSegmentContent fetches content from a URL without parsing it as M3U8
func (s *Service) fetchSegmentContent(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("segment URL is empty")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set user agent to avoid potential blocking
	req.Header.Set("User-Agent", "youtube-curator/1.0")

	// Fetch segment content
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch segment content: %w", err)
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

	// Parse the segment content directly as VTT/SRT (not M3U8)
	segmentText := s.parseVTTOrSRTContent(string(body))

	if segmentText == "" {
		return "", fmt.Errorf("no text content found in segment")
	}

	// Optimize text for token efficiency
	optimizedText := s.optimizeSubtitleText(segmentText)

	return optimizedText, nil
}

// parseVTTOrSRTContent parses only VTT or SRT content (not M3U8)
func (s *Service) parseVTTOrSRTContent(content string) string {
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

// JSONSubtitleSegment represents a segment in JSON subtitle format
type JSONSubtitleSegment struct {
	UTF8      string `json:"utf8"`
	TOffsetMs int    `json:"tOffsetMs"`
	AcAsrConf int    `json:"acAsrConf"`
}

// JSONSubtitleEvent represents an event in JSON subtitle format
type JSONSubtitleEvent struct {
	TStartMs    int                   `json:"tStartMs"`
	DDurationMs int                   `json:"dDurationMs"`
	WwinId      int                   `json:"wWinId"`
	AAppend     int                   `json:"aAppend"`
	Segs        []JSONSubtitleSegment `json:"segs"`
}

// JSONSubtitleRoot represents the root of JSON subtitle format
type JSONSubtitleRoot struct {
	Events []JSONSubtitleEvent `json:"events"`
}

// optimizeSubtitleText performs comprehensive optimization of subtitle text
func (s *Service) optimizeSubtitleText(text string) string {
	// First try to parse as JSON format (YouTube's new format)
	if strings.HasPrefix(strings.TrimSpace(text), "{") {
		if jsonText := s.parseJSONSubtitles(text); jsonText != "" {
			text = jsonText
		}
	}

	// Remove common filler words and patterns
	text = s.removeFillerWords(text)

	// Remove consecutive duplicate sentences
	text = s.removeDuplicateSentences(text)

	// Apply length limits for very long content
	text = s.truncateIfTooLong(text)

	return strings.TrimSpace(text)
}

// parseJSONSubtitles extracts text from YouTube's JSON subtitle format
func (s *Service) parseJSONSubtitles(jsonContent string) string {
	var subtitleData JSONSubtitleRoot

	if err := json.Unmarshal([]byte(jsonContent), &subtitleData); err != nil {
		// If JSON parsing fails, return empty string to fall back to original parsing
		return ""
	}

	var textSegments []string

	// Extract text from all events, sorted by timestamp
	type TimedText struct {
		Text      string
		Timestamp int
	}

	var timedTexts []TimedText

	for _, event := range subtitleData.Events {
		for _, seg := range event.Segs {
			if seg.UTF8 != "" && seg.UTF8 != "\n" {
				// Skip music/sound effect markers
				if strings.HasPrefix(seg.UTF8, "[") && strings.HasSuffix(seg.UTF8, "]") {
					continue
				}

				timedTexts = append(timedTexts, TimedText{
					Text:      seg.UTF8,
					Timestamp: event.TStartMs + seg.TOffsetMs,
				})
			}
		}
	}

	// Sort by timestamp to maintain chronological order
	sort.Slice(timedTexts, func(i, j int) bool {
		return timedTexts[i].Timestamp < timedTexts[j].Timestamp
	})

	// Combine text segments
	for _, timed := range timedTexts {
		textSegments = append(textSegments, timed.Text)
	}

	return strings.Join(textSegments, " ")
}

// removeFillerWords removes common filler words and patterns
func (s *Service) removeFillerWords(text string) string {
	// Common filler words and patterns in auto-generated subtitles
	fillerPatterns := []struct{ pattern, replacement string }{
		{` um `, ` `}, {` uh `, ` `}, {` like `, ` `},
		{` you know `, ` `}, {` I mean `, ` `}, {` so `, ` `},
		{` well `, ` `}, {` basically `, ` `}, {` actually `, ` `},
		{`[Music]`, ``}, {`[Applause]`, ``}, {`[Laughter]`, ``},
		{`[Sound Effects]`, ``}, {`[Background Music]`, ``},
	}

	for _, pattern := range fillerPatterns {
		text = strings.ReplaceAll(text, pattern.pattern, pattern.replacement)
	}

	// Compress multiple spaces into single spaces
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, ` `)

	return text
}

// removeDuplicateSentences removes consecutive duplicate or very similar sentences
func (s *Service) removeDuplicateSentences(text string) string {
	sentences := strings.Split(text, `. `)
	if len(sentences) <= 1 {
		return text
	}

	var uniqueSentences []string
	lastSentence := ""

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)

		// Skip very short fragments
		if len(sentence) < 10 {
			continue
		}

		// Check if this sentence is very similar to the last one
		if !s.isSimilarSentence(sentence, lastSentence) {
			uniqueSentences = append(uniqueSentences, sentence)
			lastSentence = sentence
		}
	}

	return strings.Join(uniqueSentences, `. `)
}

// isSimilarSentence checks if two sentences are very similar (likely duplicates)
func (s *Service) isSimilarSentence(sentence1, sentence2 string) bool {
	if sentence1 == sentence2 {
		return true
	}

	// Check for high word overlap (simple similarity check)
	words1 := strings.Fields(strings.ToLower(sentence1))
	words2 := strings.Fields(strings.ToLower(sentence2))

	if len(words1) == 0 || len(words2) == 0 {
		return false
	}

	// Count common words
	wordCount1 := make(map[string]int)
	for _, word := range words1 {
		wordCount1[word]++
	}

	commonWords := 0
	for _, word := range words2 {
		if wordCount1[word] > 0 {
			commonWords++
			wordCount1[word]--
		}
	}

	// If more than 70% of words are common, consider similar
	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	similarity := float64(commonWords) / float64(maxLen)
	return similarity > 0.7
}

// truncateIfTooLong limits the text length for very long videos
func (s *Service) truncateIfTooLong(text string) string {
	// Rough estimation: 1 token ≈ 4 characters
	// Target max ~3000 tokens (12000 characters) for input
	maxChars := 12000

	if len(text) <= maxChars {
		return text
	}

	// For very long content, take first 40%, middle 20%, last 40%
	firstPart := text[:int(float64(maxChars)*0.4)]
	lastPart := text[len(text)-int(float64(maxChars)*0.4):]

	// Find a good break point (end of sentence) for cleaner truncation
	if lastDot := strings.LastIndex(firstPart, `. `); lastDot > len(firstPart)-100 {
		firstPart = firstPart[:lastDot+2]
	}

	if firstDot := strings.Index(lastPart, `. `); firstDot < 100 && firstDot > 0 {
		lastPart = lastPart[firstDot+2:]
	}

	return firstPart + ` [content abbreviated] ` + lastPart
}

// parseThinkingBlocks extracts <think></think> blocks from LLM response and returns separated content
func parseThinkingBlocks(response string) (thinking string, summary string) {
	// Use regex to find and extract thinking blocks
	thinkingPattern := regexp.MustCompile(`(?s)<think>(.*?)</think>`)
	
	// Extract all thinking content
	thinkingMatches := thinkingPattern.FindAllStringSubmatch(response, -1)
	var thinkingParts []string
	
	for _, match := range thinkingMatches {
		if len(match) > 1 {
			thinkingParts = append(thinkingParts, strings.TrimSpace(match[1]))
		}
	}
	
	// Remove thinking blocks from the response to get clean summary
	cleanSummary := thinkingPattern.ReplaceAllString(response, "")
	cleanSummary = strings.TrimSpace(cleanSummary)
	
	// Clean up any multiple newlines that might be left
	cleanSummary = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(cleanSummary, "\n\n")
	
	// Join thinking parts if multiple blocks exist
	thinking = strings.Join(thinkingParts, "\n\n")
	
	return thinking, cleanSummary
}
