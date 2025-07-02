package store

import (
	"encoding/json"
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v3"
)

// Configuration key constants
const (
	newsletterConfigKey = "newsletter_config"
	smtpConfigKey      = "smtp_config"
	llmConfigKey       = "llm_config"
	checkIntervalKey   = "check_interval"
	channelsKey        = "channels"
	watchedVideosKey   = "watched_videos"
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

	// SMTP configuration methods
	GetSMTPConfig() (*SMTPConfig, error)
	SetSMTPConfig(config *SMTPConfig) error

	// LLM configuration methods
	GetLLMConfig() (*LLMConfig, error)
	SetLLMConfig(config *LLMConfig) error

	// Newsletter configuration methods
	GetNewsletterConfig() (*NewsletterConfig, error)
	SetNewsletterConfig(config *NewsletterConfig) error

	// Watched state management methods
	GetWatchedVideos() ([]string, error)
	SetVideoWatched(videoID string) error
	IsVideoWatched(videoID string) (bool, error)
	
	// ToWatch state management methods
	GetToWatchVideos() ([]string, error)
	SetVideoToWatch(videoID string) error
	UnsetVideoToWatch(videoID string) error
	IsVideoToWatch(videoID string) (bool, error)
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Server         string `json:"server"`
	Port           string `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	RecipientEmail string `json:"recipient_email"`
}

// LLMConfig holds LLM configuration for video summarization
type LLMConfig struct {
	EndpointURL string `json:"endpointUrl"` // LLM API endpoint URL
	APIKey      string `json:"apiKey"`      // API key for the LLM service
	Model       string `json:"model"`       // Model name to use (e.g., "gpt-3.5-turbo")
}

// NewsletterConfig holds newsletter configuration
type NewsletterConfig struct {
	Enabled bool `json:"enabled"` // Whether the newsletter cron is enabled
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
	key := []byte(channelsKey)

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
	key := []byte(channelsKey)
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
	key := []byte(channelsKey)
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
	key := []byte(checkIntervalKey)

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
	key := []byte(checkIntervalKey)
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, []byte(interval.String()))
	})
}

// GetSMTPConfig retrieves the SMTP configuration
func (s *BadgerStore) GetSMTPConfig() (*SMTPConfig, error) {
	var config SMTPConfig
	key := []byte(smtpConfigKey)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil // No configuration yet
		}
		if err != nil {
			return fmt.Errorf("failed to get SMTP configuration: %w", err)
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})
	return &config, err
}

// SetSMTPConfig stores the SMTP configuration
func (s *BadgerStore) SetSMTPConfig(config *SMTPConfig) error {
	key := []byte(smtpConfigKey)
	return s.db.Update(func(txn *badger.Txn) error {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal SMTP configuration: %w", err)
		}
		return txn.Set(key, configBytes)
	})
}

// GetLLMConfig retrieves the LLM configuration
func (s *BadgerStore) GetLLMConfig() (*LLMConfig, error) {
	var config LLMConfig
	key := []byte(llmConfigKey)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil // No configuration yet
		}
		if err != nil {
			return fmt.Errorf("failed to get LLM configuration: %w", err)
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})

	if err != nil {
		return nil, err
	}

	// Check if we have any configuration data
	if config.EndpointURL == "" && config.APIKey == "" && config.Model == "" {
		return nil, nil // No configuration yet
	}

	return &config, nil
}

// SetLLMConfig stores the LLM configuration
func (s *BadgerStore) SetLLMConfig(config *LLMConfig) error {
	key := []byte(llmConfigKey)
	return s.db.Update(func(txn *badger.Txn) error {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal LLM configuration: %w", err)
		}
		return txn.Set(key, configBytes)
	})
}

// GetNewsletterConfig retrieves the newsletter configuration
func (s *BadgerStore) GetNewsletterConfig() (*NewsletterConfig, error) {
	var config NewsletterConfig
	key := []byte(newsletterConfigKey)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			// Default to enabled if no configuration exists
			config.Enabled = true
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get newsletter configuration: %w", err)
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})
	
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SetNewsletterConfig stores the newsletter configuration
func (s *BadgerStore) SetNewsletterConfig(config *NewsletterConfig) error {
	key := []byte(newsletterConfigKey)
	return s.db.Update(func(txn *badger.Txn) error {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal newsletter configuration: %w", err)
		}
		return txn.Set(key, configBytes)
	})
}

// GetWatchedVideos retrieves the list of all watched video IDs
func (s *BadgerStore) GetWatchedVideos() ([]string, error) {
	var watchedVideos []string
	key := []byte(watchedVideosKey)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil // No watched videos yet
		}
		if err != nil {
			return fmt.Errorf("failed to get watched videos: %w", err)
		}
		return item.Value(func(val []byte) error {
			if len(val) == 0 {
				return nil
			}
			return json.Unmarshal(val, &watchedVideos)
		})
	})
	return watchedVideos, err
}

// SetVideoWatched marks a video as watched by adding it to the watched videos list
func (s *BadgerStore) SetVideoWatched(videoID string) error {
	key := []byte(watchedVideosKey)
	return s.db.Update(func(txn *badger.Txn) error {
		var watchedVideos []string
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get existing watched videos: %w", err)
		}
		if err == nil {
			err = item.Value(func(val []byte) error {
				if len(val) == 0 {
					return nil
				}
				return json.Unmarshal(val, &watchedVideos)
			})
			if err != nil {
				return err
			}
		}
		// Check if video is already marked as watched
		for _, existingID := range watchedVideos {
			if existingID == videoID {
				return nil // Already watched, no-op
			}
		}
		// Add new watched video
		watchedVideos = append(watchedVideos, videoID)
		watchedBytes, err := json.Marshal(watchedVideos)
		if err != nil {
			return fmt.Errorf("failed to marshal watched videos: %w", err)
		}
		return txn.Set(key, watchedBytes)
	})
}

// IsVideoWatched checks if a video is marked as watched
func (s *BadgerStore) IsVideoWatched(videoID string) (bool, error) {
	watchedVideos, err := s.GetWatchedVideos()
	if err != nil {
		return false, err
	}
	for _, id := range watchedVideos {
		if id == videoID {
			return true, nil
		}
	}
	return false, nil
}

// GetToWatchVideos retrieves the list of all to-watch video IDs
func (s *BadgerStore) GetToWatchVideos() ([]string, error) {
	var toWatchVideos []string
	key := []byte("to_watch_videos")

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil // No to-watch videos yet
		}
		if err != nil {
			return fmt.Errorf("failed to get to-watch videos: %w", err)
		}
		return item.Value(func(val []byte) error {
			if len(val) == 0 {
				return nil
			}
			return json.Unmarshal(val, &toWatchVideos)
		})
	})
	return toWatchVideos, err
}

// SetVideoToWatch marks a video as to-watch by adding it to the to-watch videos list
func (s *BadgerStore) SetVideoToWatch(videoID string) error {
	key := []byte("to_watch_videos")
	return s.db.Update(func(txn *badger.Txn) error {
		var toWatchVideos []string
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get existing to-watch videos: %w", err)
		}
		if err == nil {
			err = item.Value(func(val []byte) error {
				if len(val) == 0 {
					return nil
				}
				return json.Unmarshal(val, &toWatchVideos)
			})
			if err != nil {
				return err
			}
		}
		// Check if video is already marked as to-watch
		for _, existingID := range toWatchVideos {
			if existingID == videoID {
				return nil // Already to-watch, no-op
			}
		}
		// Add new to-watch video
		toWatchVideos = append(toWatchVideos, videoID)
		toWatchBytes, err := json.Marshal(toWatchVideos)
		if err != nil {
			return fmt.Errorf("failed to marshal to-watch videos: %w", err)
		}
		return txn.Set(key, toWatchBytes)
	})
}

// UnsetVideoToWatch removes a video from the to-watch list
func (s *BadgerStore) UnsetVideoToWatch(videoID string) error {
	key := []byte("to_watch_videos")
	return s.db.Update(func(txn *badger.Txn) error {
		var toWatchVideos []string
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get existing to-watch videos: %w", err)
		}
		if err == nil {
			err = item.Value(func(val []byte) error {
				if len(val) == 0 {
					return nil
				}
				return json.Unmarshal(val, &toWatchVideos)
			})
			if err != nil {
				return err
			}
		}
		// Remove video from list
		var updatedVideos []string
		for _, existingID := range toWatchVideos {
			if existingID != videoID {
				updatedVideos = append(updatedVideos, existingID)
			}
		}
		// If nothing changed, return early
		if len(updatedVideos) == len(toWatchVideos) {
			return nil
		}
		// Save updated list
		toWatchBytes, err := json.Marshal(updatedVideos)
		if err != nil {
			return fmt.Errorf("failed to marshal to-watch videos: %w", err)
		}
		return txn.Set(key, toWatchBytes)
	})
}

// IsVideoToWatch checks if a video is marked as to-watch
func (s *BadgerStore) IsVideoToWatch(videoID string) (bool, error) {
	toWatchVideos, err := s.GetToWatchVideos()
	if err != nil {
		return false, err
	}
	for _, id := range toWatchVideos {
		if id == videoID {
			return true, nil
		}
	}
	return false, nil
}
