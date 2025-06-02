package store

import (
	"encoding/json"
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v3"
)

// Package store provides a Store interface for database operations, with both a BadgerDB-backed implementation (BadgerStore)
// and an in-memory mock implementation (MockStore) for testing. Use dependency injection to pass the Store interface
// to components, enabling easy unit testing without a real database.

// Store defines the interface for storage operations
// (You can use mockgen or write your own mock)
//
//go:generate mockgen -destination=store_mock.go -package=store . Store
type Channel struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Store interface {
	Close() error
	GetLastCheckedVideoID(channelID string) (string, error)
	SetLastCheckedVideoID(channelID, videoID string) error
	GetLastCheckedTimestamp(channelID string) (time.Time, error)
	SetLastCheckedTimestamp(channelID string, timestamp time.Time) error

	// Channel management methods
	GetChannels() ([]Channel, error)
	AddChannel(channel Channel) error
	RemoveChannel(channelID string) error

	// Configuration methods
	GetCheckInterval() (time.Duration, error)
	SetCheckInterval(interval time.Duration) error
}

// BadgerStore handles database operations
type BadgerStore struct {
	db *badger.DB
}

// NewStore creates a new Store (BadgerStore) instance
func NewStore(dbPath string) (Store, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}
	return &BadgerStore{db: db}, nil
}

// Close closes the database connection
func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// GetLastCheckedVideoID retrieves the ID of the last checked video for a channel
func (s *BadgerStore) GetLastCheckedVideoID(channelID string) (string, error) {
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
func (s *BadgerStore) SetLastCheckedVideoID(channelID, videoID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(channelID), []byte(videoID))
	})
}

// GetLastCheckedTimestamp retrieves the timestamp of the last video check for a channel
// This can be used as an alternative or in conjunction with VideoID
func (s *BadgerStore) GetLastCheckedTimestamp(channelID string) (time.Time, error) {
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
func (s *BadgerStore) SetLastCheckedTimestamp(channelID string, timestamp time.Time) error {
	key := []byte(channelID) // Use channel ID directly as key
	return s.db.Update(func(txn *badger.Txn) error {
		val, err := timestamp.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal timestamp: %w", err)
		}
		return txn.Set(key, val)
	})
}

// GetChannels retrieves the list of all configured channels
func (s *BadgerStore) GetChannels() ([]Channel, error) {
	var channels []Channel
	key := []byte("channels")

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil // No channels configured yet
		}
		if err != nil {
			return fmt.Errorf("failed to get channels: %w", err)
		}
		return item.Value(func(val []byte) error {
			if len(val) == 0 {
				return nil
			}
			return json.Unmarshal(val, &channels)
		})
	})
	return channels, err
}

// AddChannel adds a new channel to the list of configured channels
func (s *BadgerStore) AddChannel(channel Channel) error {
	key := []byte("channels")
	return s.db.Update(func(txn *badger.Txn) error {
		var channels []Channel
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get existing channels: %w", err)
		}
		if err == nil {
			err = item.Value(func(val []byte) error {
				if len(val) == 0 {
					return nil
				}
				return json.Unmarshal(val, &channels)
			})
			if err != nil {
				return err
			}
		}
		// Check if channel already exists
		for _, existing := range channels {
			if existing.ID == channel.ID {
				return nil // Channel already exists, no-op
			}
		}
		// Add new channel
		channels = append(channels, channel)
		channelsBytes, err := json.Marshal(channels)
		if err != nil {
			return fmt.Errorf("failed to marshal channels: %w", err)
		}
		return txn.Set(key, channelsBytes)
	})
}

// RemoveChannel removes a channel from the list of configured channels
func (s *BadgerStore) RemoveChannel(channelID string) error {
	key := []byte("channels")
	return s.db.Update(func(txn *badger.Txn) error {
		var channels []Channel
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil // No channels to remove
		}
		if err != nil {
			return fmt.Errorf("failed to get existing channels: %w", err)
		}
		err = item.Value(func(val []byte) error {
			if len(val) == 0 {
				return nil
			}
			return json.Unmarshal(val, &channels)
		})
		if err != nil {
			return err
		}
		// Filter out the channel to remove
		var filteredChannels []Channel
		for _, existing := range channels {
			if existing.ID != channelID {
				filteredChannels = append(filteredChannels, existing)
			}
		}
		channelsBytes, err := json.Marshal(filteredChannels)
		if err != nil {
			return fmt.Errorf("failed to marshal channels: %w", err)
		}
		return txn.Set(key, channelsBytes)
	})
}

// GetCheckInterval retrieves the configured check interval
func (s *BadgerStore) GetCheckInterval() (time.Duration, error) {
	var interval time.Duration
	key := []byte("check_interval")

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			interval = time.Hour // Default to 1 hour
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get check interval: %w", err)
		}
		return item.Value(func(val []byte) error {
			durationStr := string(val)
			parsed, err := time.ParseDuration(durationStr)
			if err != nil {
				return fmt.Errorf("failed to parse duration: %w", err)
			}
			interval = parsed
			return nil
		})
	})
	return interval, err
}

// SetCheckInterval stores the check interval configuration
func (s *BadgerStore) SetCheckInterval(interval time.Duration) error {
	key := []byte("check_interval")
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, []byte(interval.String()))
	})
}
