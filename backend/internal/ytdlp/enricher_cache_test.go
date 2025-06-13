package ytdlp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"youtube-curator-v2/internal/rss"
)

func TestCacheKey(t *testing.T) {
	enricher := NewDefaultEnricher()
	
	key1 := enricher.getCacheKey("dQw4w9WgXcQ")
	key2 := enricher.getCacheKey("dQw4w9WgXcQ")
	key3 := enricher.getCacheKey("different123")
	
	// Same video ID should generate same key
	if key1 != key2 {
		t.Errorf("Expected same cache key for same video ID, got %s and %s", key1, key2)
	}
	
	// Different video IDs should generate different keys
	if key1 == key3 {
		t.Errorf("Expected different cache keys for different video IDs, both got %s", key1)
	}
	
	// Keys should be valid filenames
	if !filepath.IsAbs("/" + key1) {
		t.Errorf("Cache key should be a valid filename, got %s", key1)
	}
}

func TestCaching(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "ytdlp-test-cache")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create enricher with cache enabled
	enricher := NewDefaultEnricher()
	enricher.cacheDir = tempDir
	enricher.enableCache = true
	
	videoID := "testVideoID123"
	
	// First check - should be cache miss
	if data, found := enricher.loadFromCache(videoID); found {
		t.Errorf("Expected cache miss for new video ID, but got data: %+v", data)
	}
	
	// Create mock data
	testData := YtdlpOutput{
		Duration: 420,
		Tags:     []string{"test", "video"},
		AutomaticCaptions: map[string][]SubtitleInfo{
			"en": {{Ext: "vtt", URL: "https://example.com/subs.vtt"}},
		},
	}
	
	// Serialize and save to cache
	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	enricher.saveToCache(videoID, jsonData)
	
	// Second check - should be cache hit
	cachedData, found := enricher.loadFromCache(videoID)
	if !found {
		t.Errorf("Expected cache hit after saving data")
	}
	
	if cachedData == nil {
		t.Errorf("Expected cached data, got nil")
	} else {
		if cachedData.Duration != testData.Duration {
			t.Errorf("Expected duration %v, got %v", testData.Duration, cachedData.Duration)
		}
		
		if len(cachedData.Tags) != len(testData.Tags) {
			t.Errorf("Expected %d tags, got %d", len(testData.Tags), len(cachedData.Tags))
		}
	}
}

func TestCacheDisabled(t *testing.T) {
	// Create enricher with cache disabled
	enricher := NewDefaultEnricher()
	enricher.enableCache = false
	
	videoID := "testVideoID123"
	
	// Should always be cache miss when disabled
	if _, found := enricher.loadFromCache(videoID); found {
		t.Errorf("Expected cache miss when caching is disabled")
	}
	
	// Saving should be no-op when disabled
	enricher.saveToCache(videoID, []byte(`{"duration": 123}`))
	
	// Should still be cache miss
	if _, found := enricher.loadFromCache(videoID); found {
		t.Errorf("Expected cache miss after save when caching is disabled")
	}
}

func TestEnrichEntryWithCachedData(t *testing.T) {
	// Create enricher with cache enabled
	enricher := NewDefaultEnricher()
	
	entry := &rss.Entry{
		ID: "yt:video:testVideoID123",
	}
	
	testData := &YtdlpOutput{
		Duration: 420,
		Tags:     []string{"test", "video"},
		Comments: []Comment{
			{Text: "Great video!", Author: "User1", LikeCount: 10},
		},
		AutomaticCaptions: map[string][]SubtitleInfo{
			"en": {{Ext: "vtt", URL: "https://example.com/subs.vtt"}},
		},
	}
	
	err := enricher.enrichEntryWithData(entry, testData)
	if err != nil {
		t.Errorf("Expected no error from enrichEntryWithData, got: %v", err)
	}
	
	// Verify entry was enriched
	if entry.Duration != 420 {
		t.Errorf("Expected duration 420, got %d", entry.Duration)
	}
	
	if len(entry.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(entry.Tags))
	}
	
	if len(entry.TopComments) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(entry.TopComments))
	}
	
	if entry.AutoSubtitles != "https://example.com/subs.vtt" {
		t.Errorf("Expected subtitle URL, got %s", entry.AutoSubtitles)
	}
}

func TestClearCache(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "ytdlp-test-cache")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create enricher with cache enabled
	enricher := NewDefaultEnricher()
	enricher.cacheDir = tempDir
	enricher.enableCache = true
	
	// Add some test cache files
	testData := []byte(`{"duration": 123}`)
	enricher.saveToCache("video1", testData)
	enricher.saveToCache("video2", testData)
	
	// Verify files exist
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory: %v", err)
	}
	
	if len(entries) != 2 {
		t.Errorf("Expected 2 cache files, got %d", len(entries))
	}
	
	// Clear cache
	err = enricher.ClearCache()
	if err != nil {
		t.Errorf("Expected no error from ClearCache, got: %v", err)
	}
	
	// Verify files are gone
	entries, err = os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory after clear: %v", err)
	}
	
	if len(entries) != 0 {
		t.Errorf("Expected 0 cache files after clear, got %d", len(entries))
	}
}