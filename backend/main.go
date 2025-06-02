package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"youtube-curator-v2/internal/api"
	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("YouTube Curator v2 Starting...")
	fmt.Printf("Loaded configuration: %+v\n", cfg)
	fmt.Printf("Checking for new videos on schedule %s\n", cfg.CronSchedule)

	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	db, err := store.NewStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer db.Close()

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

	// Create email sender
	emailSender := email.NewEmailSender(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)

	// Start API server if enabled
	if cfg.EnableAPI {
		go func() {
			fmt.Printf("Starting API server on port %s...\n", cfg.APIPort)
			e := api.SetupRouter(db, feedProvider, emailSender, cfg, channelProcessor)
			if err := e.Start(":" + cfg.APIPort); err != nil {
				log.Printf("API server error: %v", err)
			}
		}()
	}

	// Run the check immediately on startup
	// checkForNewVideos(cfg, emailSender, channelProcessor, db)

	// If DebugSkipCron is set, skip the scheduler feature
	if cfg.DebugSkipCron {
		fmt.Println("DEBUG_SKIP_CRON is set: Skipping scheduler. Exiting after one check.")
		return
	}

	// Use robfig/cron for scheduling if CronSchedule is set
	fmt.Printf("Starting cron scheduler with schedule: %s\n", cfg.CronSchedule)
	c := cron.New()
	_, err = c.AddFunc(cfg.CronSchedule, func() {
		checkForNewVideos(cfg, emailSender, channelProcessor, db)
	})
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}
	c.Start()
	select {} // Block forever
}

func checkForNewVideos(cfg *config.Config, emailSender email.Sender, channelProcessor processor.ChannelProcessor, db store.Store) {
	log.Println("Checking for new videos...")
	ctx := context.Background()

	// Get channels from database instead of config
	channels, err := db.GetChannels()
	if err != nil {
		log.Printf("Error getting channels from database: %v", err)
		return
	}

	if len(channels) == 0 {
		fmt.Println("No channels configured. Use the API to add channels.")
		return
	}

	// Map to store the latest new video for each channel
	latestNewVideoPerChannel := make(map[string]rss.Entry)

	// Process each channel using the channel processor
	for _, channel := range channels {
		result := channelProcessor.ProcessChannel(ctx, channel.ID)

		// Skip channels that had errors
		if result.Error != nil {
			log.Printf("Error processing channel %s: %v\n", channel.ID, result.Error)
			continue
		}

		// If a new video was found, add it to our map
		if result.NewVideo != nil {
			latestNewVideoPerChannel[channel.ID] = *result.NewVideo
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
