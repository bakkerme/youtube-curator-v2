package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DBPath         string
	SMTPServer     string
	SMTPPort       string
	SMTPUsername   string
	SMTPPassword   string
	RecipientEmail string
	CheckInterval  time.Duration
	APIPort        string // Port for the API server
	EnableAPI      bool   // Whether to enable the API server
	CronSchedule   string // e.g. '0 0 * * *' for daily at midnight
	RSSConcurrency int    // Number of concurrent RSS fetches, default 5

	DebugMockRSS     bool
	DebugSkipCron    bool
	DebugSkipSummary bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Attempt to load .env file, but don't error if it doesn't exist
	// This allows for configuration solely through environment variables in production
	_ = godotenv.Load()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("DB_PATH environment variable is required")
	}

	smtpServer := os.Getenv("SMTP_SERVER")
	if smtpServer == "" {
		return nil, fmt.Errorf("SMTP_SERVER environment variable is required")
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		return nil, fmt.Errorf("SMTP_PORT environment variable is required")
	}

	smtpUsername := os.Getenv("SMTP_USERNAME")
	if smtpUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME environment variable is required")
	}

	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD environment variable is required")
	}

	recipientEmail := os.Getenv("RECIPIENT_EMAIL")
	if recipientEmail == "" {
		return nil, fmt.Errorf("RECIPIENT_EMAIL environment variable is required")
	}

	checkIntervalStr := os.Getenv("CHECK_INTERVAL")
	if checkIntervalStr == "" {
		checkIntervalStr = "1h" // Default to 1 hour
	}
	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CHECK_INTERVAL: %v", err)
	}

	debugMockRSSStr := os.Getenv("DEBUG_MOCK_RSS")
	debugMockRSS := strings.ToLower(debugMockRSSStr) == "true"

	debugSkipCronStr := os.Getenv("DEBUG_SKIP_CRON")
	debugSkipCron := strings.ToLower(debugSkipCronStr) == "true"

	debugSkipSummaryStr := os.Getenv("DEBUG_SKIP_SUMMARY")
	debugSkipSummary := strings.ToLower(debugSkipSummaryStr) == "true"

	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		cronSchedule = "0 0 * * *" // Default to daily at midnight
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080" // Default API port
	}

	enableAPI := true // default to true
	enableAPIStr := os.Getenv("ENABLE_API")
	if enableAPIStr != "" {
		enableAPI = strings.ToLower(enableAPIStr) == "true"
	}

	rssConcurrency := 5 // default to 5 concurrent fetches
	rssConcurrencyStr := os.Getenv("RSS_CONCURRENCY")
	if rssConcurrencyStr != "" {
		if parsed, err := parseIntEnv("RSS_CONCURRENCY", rssConcurrencyStr); err == nil && parsed > 0 {
			rssConcurrency = parsed
		} else {
			fmt.Printf("Warning: Invalid RSS_CONCURRENCY value '%s'. Using default value: %d\n", rssConcurrencyStr, rssConcurrency)
		}
	}

	return &Config{
		DBPath:         dbPath,
		SMTPServer:     smtpServer,
		SMTPPort:       smtpPort,
		SMTPUsername:   smtpUsername,
		SMTPPassword:   smtpPassword,
		RecipientEmail: recipientEmail,
		CheckInterval:  checkInterval,
		APIPort:        apiPort,
		EnableAPI:      enableAPI,
		CronSchedule:   cronSchedule,
		RSSConcurrency: rssConcurrency,

		DebugMockRSS:     debugMockRSS,
		DebugSkipCron:    debugSkipCron,
		DebugSkipSummary: debugSkipSummary,
	}, nil
}

// parseIntEnv is a helper function to parse integer environment variables
func parseIntEnv(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value '%s': must be an integer", name, value)
	}
	return parsed, nil
}
