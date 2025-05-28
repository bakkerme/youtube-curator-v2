package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Channels       []string
	SMTPServer     string
	SMTPPort       string
	SMTPUsername   string
	SMTPPassword   string
	RecipientEmail string
	CheckInterval  time.Duration
	DebugMockRSS   bool
	DebugSkipCron  bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Attempt to load .env file, but don't error if it doesn't exist
	// This allows for configuration solely through environment variables in production
	_ = godotenv.Load()

	channelsStr := os.Getenv("CHANNELS")
	if channelsStr == "" {
		log.Fatal("CHANNELS environment variable is required")
	}

	smtpServer := os.Getenv("SMTP_SERVER")
	if smtpServer == "" {
		log.Fatal("SMTP_SERVER environment variable is required")
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		log.Fatal("SMTP_PORT environment variable is required")
	}

	smtpUsername := os.Getenv("SMTP_USERNAME")
	if smtpUsername == "" {
		log.Fatal("SMTP_USERNAME environment variable is required")
	}

	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		log.Fatal("SMTP_PASSWORD environment variable is required")
	}

	recipientEmail := os.Getenv("RECIPIENT_EMAIL")
	if recipientEmail == "" {
		log.Fatal("RECIPIENT_EMAIL environment variable is required")
	}

	checkIntervalStr := os.Getenv("CHECK_INTERVAL")
	if checkIntervalStr == "" {
		checkIntervalStr = "1h" // Default to 1 hour
	}
	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		log.Fatalf("Invalid CHECK_INTERVAL: %v", err)
	}

	debugMockRSSStr := os.Getenv("DEBUG_MOCK_RSS")
	debugMockRSS := strings.ToLower(debugMockRSSStr) == "true"

	debugSkipCronStr := os.Getenv("DEBUG_SKIP_CRON")
	debugSkipCron := strings.ToLower(debugSkipCronStr) == "true"

	return &Config{
		Channels:       parseChannels(channelsStr),
		SMTPServer:     smtpServer,
		SMTPPort:       smtpPort,
		SMTPUsername:   smtpUsername,
		SMTPPassword:   smtpPassword,
		RecipientEmail: recipientEmail,
		CheckInterval:  checkInterval,
		DebugMockRSS:   debugMockRSS,
		DebugSkipCron:  debugSkipCron,
	}
}

func parseChannels(channelsStr string) []string {
	channels := strings.Split(channelsStr, ",")
	for i, ch := range channels {
		channels[i] = strings.TrimSpace(ch)
	}
	return channels
}
