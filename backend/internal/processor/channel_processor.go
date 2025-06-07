package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
)

// ChannelResult represents the result of processing a single channel
type ChannelResult struct {
	ChannelID string
	NewVideo  *rss.Entry // nil if no new video found
	Error     error
}

// ChannelProcessor defines the interface for processing YouTube channels
type ChannelProcessor interface {
	// ProcessChannel processes a single channel and returns the latest new video (if any)
	ProcessChannel(ctx context.Context, channelID string) ChannelResult
}

// DefaultChannelProcessor implements the ChannelProcessor interface
type DefaultChannelProcessor struct {
	db           store.Store
	feedProvider rss.FeedProvider
	videoStore   *store.VideoStore
}

// NewDefaultChannelProcessor creates a new instance of DefaultChannelProcessor
func NewDefaultChannelProcessor(db store.Store, feedProvider rss.FeedProvider, videoStore *store.VideoStore) *DefaultChannelProcessor {
	return &DefaultChannelProcessor{
		db:           db,
		feedProvider: feedProvider,
		videoStore:   videoStore,
	}
}

// ProcessChannel implements ChannelProcessor.ProcessChannel
func (p *DefaultChannelProcessor) ProcessChannel(ctx context.Context, channelID string) ChannelResult {
	fmt.Printf("\nFetching RSS feed for channel ID: %s\n", channelID)

	feed, err := p.feedProvider.FetchFeed(ctx, channelID)
	if err != nil {
		log.Printf("Error fetching feed for channel ID %s: %v\n", channelID, err)
		return ChannelResult{
			ChannelID: channelID,
			Error:     fmt.Errorf("error fetching feed: %w", err),
		}
	}

	lastCheckedTimestamp, err := p.db.GetLastCheckedTimestamp(channelID)
	if err != nil {
		log.Printf("Error getting last checked timestamp for channel ID %s: %v\n", channelID, err)
		// If there's an error getting the timestamp, treat it as if no previous check occurred.
		lastCheckedTimestamp = time.Time{}
	}

	var latestVideoThisChannel *rss.Entry          // Pointer to keep track of the latest new video for the current channel
	latestTimestampThisRun := lastCheckedTimestamp // Keep track of the latest timestamp for DB update

	for _, entry := range feed.Entries {
		// Store all videos in the video store (not just new ones)
		if p.videoStore != nil {
			p.videoStore.AddVideo(channelID, entry)
		}

		// Check if the video is newer than the last checked timestamp
		if entry.Published.After(lastCheckedTimestamp) {
			// If this is the first new video found for this channel, or it's newer than the current latest
			if latestVideoThisChannel == nil || entry.Published.After(latestVideoThisChannel.Published) {
				entryCopy := entry // Make a copy to avoid pointer issues
				latestVideoThisChannel = &entryCopy
			}
			// Always update latest timestamp for DB, regardless of whether it's the video we email
			if entry.Published.After(latestTimestampThisRun) {
				latestTimestampThisRun = entry.Published
			}
		}
	}

	// If a new video was found for this channel, log it
	if latestVideoThisChannel != nil {
		fmt.Printf("Found 1 new video to potentially email from channel ID %s (latest: %s)\n", channelID, latestVideoThisChannel.Title)
	}

	// Always update the last checked timestamp for the channel in the database
	if !latestTimestampThisRun.Equal(lastCheckedTimestamp) {
		if err := p.db.SetLastCheckedTimestamp(channelID, latestTimestampThisRun); err != nil {
			log.Printf("Error setting last checked timestamp for channel ID %s: %v\n", channelID, err)
		}
		fmt.Printf("Updated last checked timestamp for channel ID %s to %s\n", channelID, latestTimestampThisRun.Format(time.RFC3339))
	} else {
		fmt.Printf("No new videos found or timestamp unchanged for channel ID %s since last check (%s)\n", channelID, lastCheckedTimestamp.Format(time.RFC3339))
	}

	return ChannelResult{
		ChannelID: channelID,
		NewVideo:  latestVideoThisChannel,
		Error:     nil,
	}
}
