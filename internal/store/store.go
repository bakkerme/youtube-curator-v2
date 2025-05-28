package store

import (
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v3"
)

// Store handles database operations
type Store struct {
	db *badger.DB
}

// NewStore creates a new Store instance
func NewStore(dbPath string) (*Store, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// GetLastCheckedVideoID retrieves the ID of the last checked video for a channel
func (s *Store) GetLastCheckedVideoID(channelID string) (string, error) {
	var videoID string
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(channelID))
		if err == badger.ErrKeyNotFound {
			return nil // No entry yet, not an error
		}
		if err != nil {
			return fmt.Errorf("failed to get last checked video ID for channel %s: %w", channelID, err)
		}
		return item.Value(func(val []byte) error {
			videoID = string(val)
			return nil
		})
	})
	return videoID, err
}

// SetLastCheckedVideoID stores the ID of the last checked video for a channel
func (s *Store) SetLastCheckedVideoID(channelID, videoID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(channelID), []byte(videoID))
	})
}

// GetLastCheckedTimestamp retrieves the timestamp of the last video check for a channel
// This can be used as an alternative or in conjunction with VideoID
func (s *Store) GetLastCheckedTimestamp(channelID string) (time.Time, error) {
	var lastChecked time.Time
	key := []byte(channelID) // Use channel ID directly as key

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil // No entry yet, return zero time
		}
		if err != nil {
			return fmt.Errorf("failed to get last checked timestamp for %s: %w", channelID, err)
		}
		return item.Value(func(val []byte) error {
			return lastChecked.UnmarshalBinary(val)
		})
	})
	return lastChecked, err
}

// SetLastCheckedTimestamp stores the timestamp of the last video check for a channel
func (s *Store) SetLastCheckedTimestamp(channelID string, timestamp time.Time) error {
	key := []byte(channelID) // Use channel ID directly as key
	return s.db.Update(func(txn *badger.Txn) error {
		val, err := timestamp.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal timestamp: %w", err)
		}
		return txn.Set(key, val)
	})
}
