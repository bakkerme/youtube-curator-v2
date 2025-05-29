package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("YouTube Curator v2 Starting...")
	fmt.Printf("Loaded configuration: %+v\n", cfg)
	fmt.Printf("Checking for new videos every %s\n", cfg.CheckInterval)

	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	db, err := store.NewStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer db.Close()

	emailSender := email.NewEmailSender(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)

	var feedProvider rss.FeedProvider
	if cfg.DebugMockRSS {
		fmt.Println("Using Mock RSS Feed Provider")
		defaultProvider := rss.NewFeedProvider()
		feedProvider = rss.NewMockFeedProvider(defaultProvider)
	} else {
		fmt.Println("Using Default RSS Feed Provider")
		feedProvider = rss.NewFeedProvider()
	}

	// Run the check immediately on startup
	checkForNewVideos(cfg, db, emailSender, feedProvider)

	// If DebugSkipCron is set, skip the cron/ticker feature
	if cfg.DebugSkipCron {
		fmt.Println("DEBUG_SKIP_CRON is set: Skipping cron/ticker feature. Exiting after one check.")
		return
	}

	// Then run on the configured interval
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			checkForNewVideos(cfg, db, emailSender, feedProvider)
		case <-sigChan:
			fmt.Println("Shutting down YouTube Curator v2...")
			return
		}
	}
}

func checkForNewVideos(cfg *config.Config, db store.Store, emailSender email.EmailSenderInterface, feedProvider rss.FeedProvider) {
	log.Println("Checking for new videos...")
	ctx := context.Background()

	// Map to store the latest new video for each channel
	latestNewVideoPerChannel := make(map[string]rss.Entry)

	for _, channelID := range cfg.Channels {
		fmt.Printf("\nFetching RSS feed for channel ID: %s\n", channelID)
		feed, err := feedProvider.FetchFeed(ctx, channelID)
		if err != nil {
			log.Printf("Error fetching feed for channel ID %s: %v\n", channelID, err)
			continue
		}

		lastCheckedTimestamp, err := db.GetLastCheckedTimestamp(channelID)
		if err != nil {
			log.Printf("Error getting last checked timestamp for channel ID %s: %v\n", channelID, err)
			// If there's an error getting the timestamp, treat it as if no previous check occurred.
			lastCheckedTimestamp = time.Time{}
		}

		var latestVideoThisChannel *rss.Entry          // Pointer to keep track of the latest new video for the current channel
		latestTimestampThisRun := lastCheckedTimestamp // Keep track of the latest timestamp for DB update

		for _, entry := range feed.Entries {
			// Check if the video is newer than the last checked timestamp
			if entry.Published.After(lastCheckedTimestamp) {
				// If this is the first new video found for this channel, or it's newer than the current latest
				if latestVideoThisChannel == nil || entry.Published.After(latestVideoThisChannel.Published) {
					latestVideoThisChannel = &entry
				}
				// Always update latest timestamp for DB, regardless of whether it's the video we email
				if entry.Published.After(latestTimestampThisRun) {
					latestTimestampThisRun = entry.Published
				}
			}
		}

		// If a new video was found for this channel, add the latest one to the overall map
		if latestVideoThisChannel != nil {
			latestNewVideoPerChannel[channelID] = *latestVideoThisChannel
			fmt.Printf("Found 1 new video to potentially email from channel ID %s (latest: %s)\n", channelID, latestVideoThisChannel.Title)
		}

		// Always update the last checked timestamp for the channel in the database
		if !latestTimestampThisRun.Equal(lastCheckedTimestamp) {
			if err := db.SetLastCheckedTimestamp(channelID, latestTimestampThisRun); err != nil {
				log.Printf("Error setting last checked timestamp for channel ID %s: %v\n", channelID, err)
			}
			fmt.Printf("Updated last checked timestamp for channel ID %s to %s\n", channelID, latestTimestampThisRun.Format(time.RFC3339))
		} else {
			fmt.Printf("No new videos found or timestamp unchanged for channel ID %s since last check (%s)\n", channelID, lastCheckedTimestamp.Format(time.RFC3339))
		}
	}

	// Collect the latest new videos from the map into a slice
	var videosToSendEmail []rss.Entry
	for _, video := range latestNewVideoPerChannel {
		videosToSendEmail = append(videosToSendEmail, video)
	}

	// Only send email if there are new videos from at least one channel
	if len(videosToSendEmail) > 0 {
		fmt.Printf("\nFound a total of %d new video(s) to email across all channels.\n", len(videosToSendEmail))

		// Sort videos by published date (optional, but nice for the email)
		sort.Slice(videosToSendEmail, func(i, j int) bool {
			return videosToSendEmail[i].Published.Before(videosToSendEmail[j].Published)
		})

		emailBody, err := email.FormatNewVideosEmail(videosToSendEmail)
		if err != nil {
			log.Printf("Error formatting combined email: %v\n", err)
		} else {
			subject := "New YouTube Videos Update"
			if err := emailSender.Send(cfg.RecipientEmail, subject, emailBody); err != nil {
				log.Printf("Error sending combined email: %v\n", err)
			} else {
				fmt.Printf("Combined email sent successfully to %s\n", cfg.RecipientEmail)
			}
		}
	} else {
		fmt.Println("No new videos found across all channels since last check.")
	}

	log.Println("Finished checking for new videos.")
}
