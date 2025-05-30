package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/processor"
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

	// Create the channel processor
	channelProcessor := processor.NewDefaultChannelProcessor(db, feedProvider)

	// Run the check immediately on startup
	checkForNewVideos(cfg, emailSender, channelProcessor)

	// If DebugSkipCron is set, skip the cron/ticker feature
	if cfg.DebugSkipCron {
		fmt.Println("DEBUG_SKIP_CRON is set: Skipping cron/ticker feature. Exiting after one check.")
		return
	}
}

func checkForNewVideos(cfg *config.Config, emailSender email.Sender, channelProcessor processor.ChannelProcessor) {
	log.Println("Checking for new videos...")
	ctx := context.Background()

	// Map to store the latest new video for each channel
	latestNewVideoPerChannel := make(map[string]rss.Entry)

	// Process each channel using the channel processor
	for _, channelID := range cfg.Channels {
		result := channelProcessor.ProcessChannel(ctx, channelID)

		// Skip channels that had errors
		if result.Error != nil {
			log.Printf("Error processing channel %s: %v\n", channelID, result.Error)
			continue
		}

		// If a new video was found, add it to our map
		if result.NewVideo != nil {
			latestNewVideoPerChannel[channelID] = *result.NewVideo
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
