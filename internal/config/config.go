package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DBPath         string
	Channels       []string // This will be deprecated in favor of database storage
	SMTPServer     string
	SMTPPort       string
	SMTPUsername   string
	SMTPPassword   string
	RecipientEmail string
	CheckInterval  time.Duration
	DebugMockRSS   bool
	DebugSkipCron  bool
	APIPort        string // Port for the API server
	EnableAPI      bool   // Whether to enable the API server
	ScheduleMode   string // 'interval' or 'daily'
	DailyRunTime   string // e.g. '07:00' for 7:00 AM
	CronSchedule   string // e.g. '0 0 * * *' for daily at midnight
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

	channelsFile := os.Getenv("CHANNELS_FILE")
	var channels []string
	if channelsFile != "" {
		var err error
		channels, err = loadChannelsFromFile(channelsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load channels from file: %v", err)
		}
	} else {
		return nil, fmt.Errorf("CHANNELS_FILE environment variable is required")
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

	scheduleMode := os.Getenv("SCHEDULE_MODE")
	if scheduleMode == "" {
		scheduleMode = "interval"
	}
	dailyRunTime := os.Getenv("DAILY_RUN_TIME") // e.g. '07:00'

	debugMockRSSStr := os.Getenv("DEBUG_MOCK_RSS")
	debugMockRSS := strings.ToLower(debugMockRSSStr) == "true"

	debugSkipCronStr := os.Getenv("DEBUG_SKIP_CRON")
	debugSkipCron := strings.ToLower(debugSkipCronStr) == "true"

	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		cronSchedule = "0 0 * * *" // Default to daily at midnight
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080" // Default API port
	}

	enableAPI := strings.ToLower(os.Getenv("ENABLE_API")) == "true"

	return &Config{
		DBPath:         dbPath,
		Channels:       channels,
		SMTPServer:     smtpServer,
		SMTPPort:       smtpPort,
		SMTPUsername:   smtpUsername,
		SMTPPassword:   smtpPassword,
		RecipientEmail: recipientEmail,
		CheckInterval:  checkInterval,
		DebugMockRSS:   debugMockRSS,
		DebugSkipCron:  debugSkipCron,
		APIPort:        apiPort,
		EnableAPI:      enableAPI,
		ScheduleMode:   scheduleMode,
		DailyRunTime:   dailyRunTime,
		CronSchedule:   cronSchedule,
	}, nil
}

func loadChannelsFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var channels []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		channels = append(channels, line)
	}
	return channels, nil
}
