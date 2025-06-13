package ytdlp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"youtube-curator-v2/internal/rss"
)

// CommandExecutor defines the interface for executing commands
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)
}

// DefaultCommandExecutor implements CommandExecutor using os/exec
type DefaultCommandExecutor struct{}

// Execute runs a command and returns its output
func (e *DefaultCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	log.Println("Executing command:", cmd.String())
	return cmd.Output()
}

// Enricher provides video enrichment using yt-dlp
type Enricher interface {
	EnrichEntry(ctx context.Context, entry *rss.Entry) error
	ResolveChannelID(ctx context.Context, url string) (string, error)
}

// DefaultEnricher implements Enricher using yt-dlp command
type DefaultEnricher struct {
	ytdlpPath  string
	timeout    time.Duration
	maxRetries int
	executor   CommandExecutor
	cacheDir   string
	enableCache bool
}

// NewDefaultEnricher creates a new yt-dlp enricher
func NewDefaultEnricher() *DefaultEnricher {
	// Set up cache directory
	cacheDir := os.Getenv("YTDLP_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "./cache/ytdlp" // Default cache directory
	}
	
	// Enable cache by default in development (can be disabled with env var)
	enableCache := os.Getenv("YTDLP_DISABLE_CACHE") != "true"
	
	enricher := &DefaultEnricher{
		ytdlpPath:   "yt-dlp",         // assumes yt-dlp is in PATH
		timeout:     60 * time.Second, // Increased from 30s to 60s
		maxRetries:  2,                // Allow 2 retries on failure
		executor:    &DefaultCommandExecutor{},
		cacheDir:    cacheDir,
		enableCache: enableCache,
	}
	
	// Create cache directory if caching is enabled
	if enableCache {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			log.Printf("Warning: Failed to create yt-dlp cache directory %s: %v", cacheDir, err)
			enricher.enableCache = false
		} else {
			log.Printf("yt-dlp caching enabled, using directory: %s", cacheDir)
		}
	}
	
	return enricher
}

// NewDefaultEnricherWithTimeout creates a new yt-dlp enricher with custom timeout
func NewDefaultEnricherWithTimeout(timeout time.Duration) *DefaultEnricher {
	enricher := NewDefaultEnricher()
	enricher.timeout = timeout
	return enricher
}

// NewDefaultEnricherWithConfig creates a new yt-dlp enricher with custom configuration
func NewDefaultEnricherWithConfig(timeout time.Duration, maxRetries int) *DefaultEnricher {
	enricher := NewDefaultEnricher()
	enricher.timeout = timeout
	enricher.maxRetries = maxRetries
	return enricher
}

// NewDefaultEnricherWithExecutor creates a new yt-dlp enricher with custom command executor (for testing)
func NewDefaultEnricherWithExecutor(executor CommandExecutor) *DefaultEnricher {
	enricher := NewDefaultEnricher()
	enricher.executor = executor
	return enricher
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

	// Try to load from cache first
	if cachedData, found := e.loadFromCache(videoID); found {
		// Use cached data
		return e.enrichEntryWithData(entry, cachedData)
	}

	// Cache miss, fetch from yt-dlp with retries
	var lastErr error
	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		if attempt > 0 {
			// Add a small delay between retries
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		// Create context with timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, e.timeout)

		// Build yt-dlp command arguments
		args := []string{
			"--skip-download",
			"--dump-json",
			// "--write-comments",
			"--write-auto-subs",
			"--sub-langs", "en",
			fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		}

		// Execute command using the executor interface
		output, err := e.executor.Execute(attemptCtx, e.ytdlpPath, args...)
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

		// Save raw output to cache before parsing
		e.saveToCache(videoID, output)

		// Parse JSON output
		var ytdlpData YtdlpOutput
		if err := json.Unmarshal(output, &ytdlpData); err != nil {
			return fmt.Errorf("failed to parse yt-dlp output for video %s: %w", videoID, err)
		}

		// Enrich the entry with the fetched data
		return e.enrichEntryWithData(entry, &ytdlpData)
	}

	// If we get here, all retries failed
	return lastErr
}

// enrichEntryWithData enriches an RSS entry with yt-dlp data
func (e *DefaultEnricher) enrichEntryWithData(entry *rss.Entry, ytdlpData *YtdlpOutput) error {
	// Set duration
	if ytdlpData.Duration > 0 {
		entry.Duration = int(ytdlpData.Duration)
	}

	// Set tags
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

	return nil
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

// ResolveChannelID resolves a YouTube URL (including @username, /c/, /user/ formats) to a channel ID using yt-dlp
func (e *DefaultEnricher) ResolveChannelID(ctx context.Context, url string) (string, error) {
	// Create context with timeout
	resolveCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Build yt-dlp command to get channel information
	args := []string{
		"--dump-json",
		"--playlist-items", "0", // Don't download any videos, just get channel info
		url,
	}

	// Execute command
	output, err := e.executor.Execute(resolveCtx, e.ytdlpPath, args...)
	if err != nil {
		return "", fmt.Errorf("failed to resolve channel URL %s: %w", url, err)
	}

	// Parse JSON output to extract channel ID
	var ytdlpData struct {
		ChannelID string `json:"channel_id"`
		Channel   string `json:"channel"`
	}

	if err := json.Unmarshal(output, &ytdlpData); err != nil {
		return "", fmt.Errorf("failed to parse yt-dlp output for URL %s: %w", url, err)
	}

	if ytdlpData.ChannelID == "" {
		return "", fmt.Errorf("no channel ID found for URL %s", url)
	}

	return ytdlpData.ChannelID, nil
}

// getCacheKey generates a cache key for a video ID
func (e *DefaultEnricher) getCacheKey(videoID string) string {
	// Create a hash of the video ID for a clean filename
	hash := sha256.Sum256([]byte(videoID))
	return hex.EncodeToString(hash[:])[:16] + ".json" // Use first 16 chars of hash
}

// getCachePath returns the full path to the cache file for a video ID
func (e *DefaultEnricher) getCachePath(videoID string) string {
	return filepath.Join(e.cacheDir, e.getCacheKey(videoID))
}

// loadFromCache attempts to load cached yt-dlp data for a video ID
func (e *DefaultEnricher) loadFromCache(videoID string) (*YtdlpOutput, bool) {
	if !e.enableCache {
		return nil, false
	}

	cachePath := e.getCachePath(videoID)
	
	// Check if cache file exists and is readable
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false // Cache miss or read error
	}

	// Try to parse the cached JSON
	var ytdlpData YtdlpOutput
	if err := json.Unmarshal(data, &ytdlpData); err != nil {
		log.Printf("Warning: Failed to parse cached data for video %s, removing invalid cache file", videoID)
		os.Remove(cachePath) // Remove corrupted cache file
		return nil, false
	}

	log.Printf("Cache hit for video %s", videoID)
	return &ytdlpData, true
}

// saveToCache saves yt-dlp data to cache for a video ID
func (e *DefaultEnricher) saveToCache(videoID string, data []byte) {
	if !e.enableCache {
		return
	}

	cachePath := e.getCachePath(videoID)
	
	// Write to a temporary file first, then rename for atomic operation
	tmpPath := cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		log.Printf("Warning: Failed to write cache file for video %s: %v", videoID, err)
		return
	}

	if err := os.Rename(tmpPath, cachePath); err != nil {
		log.Printf("Warning: Failed to rename cache file for video %s: %v", videoID, err)
		os.Remove(tmpPath) // Clean up tmp file
		return
	}

	log.Printf("Cached data for video %s", videoID)
}

// ClearCache removes all cached files (useful for development)
func (e *DefaultEnricher) ClearCache() error {
	if !e.enableCache {
		return nil
	}

	entries, err := os.ReadDir(e.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(e.cacheDir, entry.Name())
			if err := os.Remove(path); err != nil {
				log.Printf("Warning: Failed to remove cache file %s: %v", path, err)
			}
		}
	}

	log.Printf("Cleared yt-dlp cache directory: %s", e.cacheDir)
	return nil
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
