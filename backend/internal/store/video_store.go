package store

import (
	"fmt"
	"sync"
	"time"
	"youtube-curator-v2/internal/rss"
)

// VideoEntry represents a video with metadata and TTL
type VideoEntry struct {
	Entry     rss.Entry `json:"entry"`
	ChannelID string    `json:"channelId"`
	CachedAt  time.Time `json:"cachedAt"`
	Watched   bool      `json:"watched"`
}

// VideoStore provides in-memory storage for videos with TTL
type VideoStore struct {
	videos          map[string]VideoEntry // key: video ID
	mutex           sync.RWMutex
	ttl             time.Duration
	lastRefreshedAt time.Time
	store           Store // Reference to persistent store for watched state
}

// NewVideoStore creates a new in-memory video store with the specified TTL
func NewVideoStore(ttl time.Duration) *VideoStore {
	store := &VideoStore{
		videos:          make(map[string]VideoEntry),
		ttl:             ttl,
		lastRefreshedAt: time.Time{},
		store:           nil, // Will be set later via SetStore method
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// NewVideoStoreWithStore creates a new in-memory video store with persistent store reference
func NewVideoStoreWithStore(ttl time.Duration, store Store) *VideoStore {
	vs := &VideoStore{
		videos:          make(map[string]VideoEntry),
		ttl:             ttl,
		lastRefreshedAt: time.Time{},
		store:           store,
	}

	// Start cleanup goroutine
	go vs.cleanupExpired()

	return vs
}

// SetStore sets the persistent store reference for watched state persistence
func (vs *VideoStore) SetStore(store Store) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	vs.store = store
}

// AddVideo adds or updates a video in the store
func (vs *VideoStore) AddVideo(channelID string, entry rss.Entry) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if video already exists and preserve its watched state
	var watched bool = false
	if existingVideo, exists := vs.videos[entry.ID]; exists {
		watched = existingVideo.Watched
	} else if vs.store != nil {
		// If video doesn't exist in memory, check persistent store for watched state
		isWatched, err := vs.store.IsVideoWatched(entry.ID)
		if err != nil {
			return fmt.Errorf("failed to check watched state for video %s: %w", entry.ID, err)
		}
		watched = isWatched
	}

	vs.videos[entry.ID] = VideoEntry{
		Entry:     entry,
		ChannelID: channelID,
		CachedAt:  time.Now(),
		Watched:   watched,
	}
	return nil
}

// GetAllVideos returns all non-expired videos
func (vs *VideoStore) GetAllVideos() []VideoEntry {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	now := time.Now()
	var validVideos []VideoEntry

	for _, video := range vs.videos {
		if now.Sub(video.CachedAt) < vs.ttl {
			validVideos = append(validVideos, video)
		}
	}

	return validVideos
}

// cleanupExpired runs periodically to remove expired videos
func (vs *VideoStore) cleanupExpired() {
	ticker := time.NewTicker(time.Hour) // Run cleanup every hour
	defer ticker.Stop()

	for range ticker.C {
		vs.mutex.Lock()
		now := time.Now()

		for id, video := range vs.videos {
			if now.Sub(video.CachedAt) >= vs.ttl {
				delete(vs.videos, id)
			}
		}
		vs.mutex.Unlock()
	}
}

// GetVideoCount returns the number of videos currently in the store
func (vs *VideoStore) GetVideoCount() int {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()
	return len(vs.videos)
}

// SetLastRefreshedAt sets the last refreshed timestamp
func (vs *VideoStore) SetLastRefreshedAt(t time.Time) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	vs.lastRefreshedAt = t
}

// GetLastRefreshedAt returns the last refreshed timestamp
func (vs *VideoStore) GetLastRefreshedAt() time.Time {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()
	return vs.lastRefreshedAt
}

// MarkVideoAsWatched sets the Watched flag to true for the video with the given ID
func (vs *VideoStore) MarkVideoAsWatched(videoID string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	if video, ok := vs.videos[videoID]; ok {
		video.Watched = true
		vs.videos[videoID] = video
	}

	// Persist watched state to database if store is available
	if vs.store != nil {
		if err := vs.store.SetVideoWatched(videoID); err != nil {
			return fmt.Errorf("failed to persist watched state for video %s: %w", videoID, err)
		}
	}
	return nil
}
