package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

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

	// Create video store with 24 hour TTL
	videoStore := store.NewVideoStore(24 * time.Hour)

	// Create the channel processor
	channelProcessor := processor.NewDefaultChannelProcessor(db, feedProvider, videoStore)

	// Create email sender with SMTP settings from database
	smtpConfig, err := db.GetSMTPConfig()
	if err != nil {
		log.Printf("Warning: Failed to get SMTP configuration: %v", err)
	}

	var emailSender email.Sender
	if smtpConfig != nil && smtpConfig.Server != "" {
		emailSender = email.NewEmailSender(smtpConfig.Server, smtpConfig.Port, smtpConfig.Username, smtpConfig.Password)
	} else {
		// Fall back to environment variables for backward compatibility
		emailSender = email.NewEmailSender(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	}

	// Start API server if enabled
	if cfg.EnableAPI {
		go func() {
			fmt.Printf("Starting API server on port %s...\n", cfg.APIPort)
			e := api.SetupRouter(db, feedProvider, emailSender, cfg, channelProcessor, videoStore)
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
			// Get SMTP config from database for recipient email
			smtpConfig, err := db.GetSMTPConfig()
			if err != nil || smtpConfig == nil || smtpConfig.RecipientEmail == "" {
				// Fall back to environment variable
				recipientEmail := cfg.RecipientEmail
				if recipientEmail == "" {
					log.Printf("Error: No recipient email configured\n")
					return
				}
				subject := "New YouTube Videos Update"
				if err := emailSender.Send(recipientEmail, subject, emailBody); err != nil {
					log.Printf("Error sending combined email: %v\n", err)
				} else {
					fmt.Printf("Combined email sent successfully to %s\n", recipientEmail)
				}
			} else {
				subject := "New YouTube Videos Update"
				if err := emailSender.Send(smtpConfig.RecipientEmail, subject, emailBody); err != nil {
					log.Printf("Error sending combined email: %v\n", err)
				} else {
					fmt.Printf("Combined email sent successfully to %s\n", smtpConfig.RecipientEmail)
				}
			}
		}
	} else {
		fmt.Println("No new videos found across all channels since last check.")
	}

	log.Println("Finished checking for new videos.")
}
