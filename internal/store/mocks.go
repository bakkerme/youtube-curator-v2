package store

import (
	"fmt"
	"time"
)

// MockStore is an in-memory implementation of the Store interface for testing
// Not concurrency-safe, intended for unit tests only
type MockStore struct {
	videoIDs   map[string]string
	timestamps map[string]time.Time
	closed     bool
}

func NewMockStore() *MockStore {
	return &MockStore{
		videoIDs:   make(map[string]string),
		timestamps: make(map[string]time.Time),
	}
}

func (m *MockStore) Close() error {
	m.closed = true
	return nil
}

func (m *MockStore) GetLastCheckedVideoID(channelID string) (string, error) {
	if m.closed {
		return "", fmt.Errorf("store is closed")
	}
	return m.videoIDs[channelID], nil
}

func (m *MockStore) SetLastCheckedVideoID(channelID, videoID string) error {
	if m.closed {
		return fmt.Errorf("store is closed")
	}
	m.videoIDs[channelID] = videoID
	return nil
}

func (m *MockStore) GetLastCheckedTimestamp(channelID string) (time.Time, error) {
	if m.closed {
		return time.Time{}, fmt.Errorf("store is closed")
	}
	ts, ok := m.timestamps[channelID]
	if !ok {
		return time.Time{}, nil
	}
	return ts, nil
}

func (m *MockStore) SetLastCheckedTimestamp(channelID string, timestamp time.Time) error {
	if m.closed {
		return fmt.Errorf("store is closed")
	}
	m.timestamps[channelID] = timestamp
	return nil
}
