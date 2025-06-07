package store

import (
	"sync"
	"time"
	"youtube-curator-v2/internal/rss"
)

// VideoEntry represents a video with metadata and TTL
type VideoEntry struct {
	Entry     rss.Entry `json:"entry"`
	ChannelID string    `json:"channelId"`
	CachedAt  time.Time `json:"cachedAt"`
}

// VideoStore provides in-memory storage for videos with TTL
type VideoStore struct {
	videos          map[string]VideoEntry // key: video ID
	mutex           sync.RWMutex
	ttl             time.Duration
	lastRefreshedAt time.Time
}

// NewVideoStore creates a new in-memory video store with the specified TTL
func NewVideoStore(ttl time.Duration) *VideoStore {
	store := &VideoStore{
		videos:          make(map[string]VideoEntry),
		ttl:             ttl,
		lastRefreshedAt: time.Time{},
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// AddVideo adds or updates a video in the store
func (vs *VideoStore) AddVideo(channelID string, entry rss.Entry) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	vs.videos[entry.ID] = VideoEntry{
		Entry:     entry,
		ChannelID: channelID,
		CachedAt:  time.Now(),
	}
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
