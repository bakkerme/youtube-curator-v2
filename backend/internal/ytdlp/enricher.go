package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"youtube-curator-v2/internal/rss"
)

// Enricher provides video enrichment using yt-dlp
type Enricher interface {
	EnrichEntry(ctx context.Context, entry *rss.Entry) error
}

// DefaultEnricher implements Enricher using yt-dlp command
type DefaultEnricher struct {
	ytdlpPath  string
	timeout    time.Duration
	maxRetries int
}

// NewDefaultEnricher creates a new yt-dlp enricher
func NewDefaultEnricher() *DefaultEnricher {
	return &DefaultEnricher{
		ytdlpPath:  "yt-dlp",         // assumes yt-dlp is in PATH
		timeout:    60 * time.Second, // Increased from 30s to 60s
		maxRetries: 2,                // Allow 2 retries on failure
	}
}

// NewDefaultEnricherWithTimeout creates a new yt-dlp enricher with custom timeout
func NewDefaultEnricherWithTimeout(timeout time.Duration) *DefaultEnricher {
	return &DefaultEnricher{
		ytdlpPath:  "yt-dlp",
		timeout:    timeout,
		maxRetries: 2,
	}
}

// NewDefaultEnricherWithConfig creates a new yt-dlp enricher with custom configuration
func NewDefaultEnricherWithConfig(timeout time.Duration, maxRetries int) *DefaultEnricher {
	return &DefaultEnricher{
		ytdlpPath:  "yt-dlp",
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

// YtdlpOutput represents the JSON output structure from yt-dlp
type YtdlpOutput struct {
	Duration          float64                   `json:"duration"`
	Tags              []string                  `json:"tags"`
	Subtitles         map[string][]SubtitleInfo `json:"subtitles"`
	AutomaticCaptions map[string][]SubtitleInfo `json:"automatic_captions"`
	Comments          []Comment                 `json:"comments"`
}

// SubtitleInfo represents subtitle information
type SubtitleInfo struct {
	Ext string `json:"ext"`
	URL string `json:"url"`
}

// Comment represents a video comment
type Comment struct {
	Text      string `json:"text"`
	Author    string `json:"author"`
	LikeCount int    `json:"like_count"`
}

// EnrichEntry enriches an RSS entry with yt-dlp data
func (e *DefaultEnricher) EnrichEntry(ctx context.Context, entry *rss.Entry) error {
	// Extract video ID from entry
	videoID, err := extractVideoID(entry.ID)
	if err != nil {
		return fmt.Errorf("failed to extract video ID: %w", err)
	}

	fmt.Println("Enriching entry", entry.ID)

	// Try with retries
	var lastErr error
	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		if attempt > 0 {
			// Add a small delay between retries
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		// Create context with timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, e.timeout)

		// Build yt-dlp command
		cmd := exec.CommandContext(attemptCtx, e.ytdlpPath,
			"--skip-download",
			"--dump-json",
			"--write-comments",
			"--write-auto-subs",
			"--sub-langs", "en",
			fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		)
		fmt.Println(cmd.String())

		// Capture both stdout and stderr for better error reporting
		output, err := cmd.Output()
		cancel() // Clean up context immediately after command

		if err != nil {
			// Try to get stderr for more detailed error information
			if exitError, ok := err.(*exec.ExitError); ok {
				stderr := string(exitError.Stderr)
				if stderr != "" {
					lastErr = fmt.Errorf("yt-dlp command failed (exit status %d): %s", exitError.ExitCode(), stderr)
				} else {
					lastErr = fmt.Errorf("yt-dlp command failed for video %s: %w", videoID, err)
				}
			} else if attemptCtx.Err() == context.DeadlineExceeded {
				lastErr = fmt.Errorf("yt-dlp command timed out after %v for video %s", e.timeout, videoID)
			} else {
				lastErr = fmt.Errorf("yt-dlp command failed for video %s: %w", videoID, err)
			}

			// If this was the last attempt or a non-retryable error, return
			if attempt == e.maxRetries || !isRetryableError(err) {
				return lastErr
			}
			continue
		}

		// Parse JSON output
		var ytdlpData YtdlpOutput
		if err := json.Unmarshal(output, &ytdlpData); err != nil {
			return fmt.Errorf("failed to parse yt-dlp output for video %s: %w", videoID, err)
		}

		// Enrich the entry - success case
		if ytdlpData.Duration > 0 {
			entry.Duration = int(ytdlpData.Duration)
		}

		if len(ytdlpData.Tags) > 0 {
			entry.Tags = ytdlpData.Tags
		}

		// Extract top comments (limit to 5 for email)
		if len(ytdlpData.Comments) > 0 {
			topComments := make([]string, 0, 5)
			for i, comment := range ytdlpData.Comments {
				if i >= 5 {
					break
				}
				topComments = append(topComments, comment.Text)
			}
			entry.TopComments = topComments
		}

		// Extract auto-generated English subtitles
		if autoSubs, exists := ytdlpData.AutomaticCaptions["en"]; exists && len(autoSubs) > 0 {
			// For now, just store the first subtitle URL - we could fetch and parse it later
			entry.AutoSubtitles = autoSubs[0].URL
		}

		return nil // Success!
	}

	// If we get here, all retries failed
	return lastErr
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Retry on timeout errors and certain network-related errors
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "temporary failure")
}

// extractVideoID extracts YouTube video ID from entry ID
// Entry ID format is typically "yt:video:VIDEO_ID"
func extractVideoID(entryID string) (string, error) {
	parts := strings.Split(entryID, ":")
	if len(parts) != 3 || parts[0] != "yt" || parts[1] != "video" {
		return "", fmt.Errorf("invalid entry ID format: %s", entryID)
	}
	return parts[2], nil
}
