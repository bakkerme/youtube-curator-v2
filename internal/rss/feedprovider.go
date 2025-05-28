package rss

import (
	"context"
	"fmt"
)

// FeedProvider defines the interface for fetching and processing RSS feed data.
type FeedProvider interface {
	// FetchFeed retrieves and processes a feed for the given channel ID
	FetchFeed(ctx context.Context, channelID string) (*Feed, error)
}

// DefaultFeedProvider implements the FeedProvider interface using the standard RSS functions
type DefaultFeedProvider struct {
	// Can add configuration options here if needed
}

// NewFeedProvider creates a new instance of the default feed provider
func NewFeedProvider() *DefaultFeedProvider {
	return &DefaultFeedProvider{}
}

// FetchFeed implements FeedProvider.FetchFeed
func (p *DefaultFeedProvider) FetchFeed(ctx context.Context, channelID string) (*Feed, error) {
	url := fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID)
	rssString, err := fetchRSS(url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch RSS for channel ID %s: %w", channelID, err)
	}

	feed := &Feed{}
	err = processRSSFeed(rssString, feed)
	if err != nil {
		return nil, fmt.Errorf("could not process RSS feed from %s: %w", url, err)
	}

	return feed, nil
}
