package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"DB_PATH", "SMTP_SERVER", "SMTP_PORT", "SMTP_USERNAME",
		"SMTP_PASSWORD", "RECIPIENT_EMAIL", "CHECK_INTERVAL",
		"DEBUG_MOCK_RSS", "DEBUG_SKIP_CRON", "API_PORT", "ENABLE_API", "CRON_SCHEDULE",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Clean up function to restore environment
	cleanup := func() {
		for _, key := range envVars {
			if val, exists := originalEnv[key]; exists {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}
	defer cleanup()

	t.Run("ValidConfig", func(t *testing.T) {
		// Clear environment first
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		// Set required environment variables
		os.Setenv("DB_PATH", "./youtubecurator.db")
		os.Setenv("SMTP_SERVER", "smtp.example.com")
		os.Setenv("SMTP_PORT", "587")
		os.Setenv("SMTP_USERNAME", "user@example.com")
		os.Setenv("SMTP_PASSWORD", "password123")
		os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
		os.Setenv("CHECK_INTERVAL", "30m")
		os.Setenv("DEBUG_MOCK_RSS", "true")
		os.Setenv("DEBUG_SKIP_CRON", "false")
		os.Setenv("API_PORT", "9090")
		os.Setenv("ENABLE_API", "true")
		os.Setenv("CRON_SCHEDULE", "0 0 * * *")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if config.DBPath != "./youtubecurator.db" {
			t.Errorf("Expected DBPath to be './youtubecurator.db', got '%s'", config.DBPath)
		}
		if config.SMTPServer != "smtp.example.com" {
			t.Errorf("Expected SMTPServer to be 'smtp.example.com', got '%s'", config.SMTPServer)
		}
		if config.SMTPPort != "587" {
			t.Errorf("Expected SMTPPort to be '587', got '%s'", config.SMTPPort)
		}
		if config.SMTPUsername != "user@example.com" {
			t.Errorf("Expected SMTPUsername to be 'user@example.com', got '%s'", config.SMTPUsername)
		}
		if config.SMTPPassword != "password123" {
			t.Errorf("Expected SMTPPassword to be 'password123', got '%s'", config.SMTPPassword)
		}
		if config.RecipientEmail != "recipient@example.com" {
			t.Errorf("Expected RecipientEmail to be 'recipient@example.com', got '%s'", config.RecipientEmail)
		}
		if config.CheckInterval != 30*time.Minute {
			t.Errorf("Expected CheckInterval to be 30m, got %v", config.CheckInterval)
		}
		if !config.DebugMockRSS {
			t.Error("Expected DebugMockRSS to be true")
		}
		if config.DebugSkipCron {
			t.Error("Expected DebugSkipCron to be false")
		}
		if config.APIPort != "9090" {
			t.Errorf("Expected APIPort to be '9090', got '%s'", config.APIPort)
		}
		if !config.EnableAPI {
			t.Error("Expected EnableAPI to be true")
		}
		if config.CronSchedule != "0 0 * * *" {
			t.Errorf("Expected CronSchedule to be '0 0 * * *', got '%s'", config.CronSchedule)
		}
	})

	t.Run("DefaultValues", func(t *testing.T) {
		// Clear environment first
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		// Set only required environment variables
		os.Setenv("DB_PATH", "./youtubecurator.db")
		os.Setenv("SMTP_SERVER", "smtp.example.com")
		os.Setenv("SMTP_PORT", "587")
		os.Setenv("SMTP_USERNAME", "user@example.com")
		os.Setenv("SMTP_PASSWORD", "password123")
		os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Test default values
		if config.CheckInterval != time.Hour {
			t.Errorf("Expected default CheckInterval to be 1h, got %v", config.CheckInterval)
		}
		if config.DebugMockRSS {
			t.Error("Expected DebugMockRSS to be false by default")
		}
		if config.DebugSkipCron {
			t.Error("Expected DebugSkipCron to be false by default")
		}
		if config.APIPort != "8080" {
			t.Errorf("Expected default APIPort to be '8080', got '%s'", config.APIPort)
		}
		if config.EnableAPI {
			t.Error("Expected EnableAPI to be false by default")
		}
		if config.CronSchedule != "0 0 * * *" {
			t.Errorf("Expected default CronSchedule to be '0 0 * * *', got '%s'", config.CronSchedule)
		}
	})

	t.Run("DebugFlagsCaseInsensitive", func(t *testing.T) {
		// Clear environment first
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		// Set required environment variables with mixed case debug flags
		os.Setenv("DB_PATH", "./youtubecurator.db")
		os.Setenv("SMTP_SERVER", "smtp.example.com")
		os.Setenv("SMTP_PORT", "587")
		os.Setenv("SMTP_USERNAME", "user@example.com")
		os.Setenv("SMTP_PASSWORD", "password123")
		os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
		os.Setenv("DEBUG_MOCK_RSS", "TRUE")
		os.Setenv("DEBUG_SKIP_CRON", "True")
		os.Setenv("ENABLE_API", "True")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !config.DebugMockRSS {
			t.Error("Expected DebugMockRSS to be true with 'TRUE'")
		}
		if !config.DebugSkipCron {
			t.Error("Expected DebugSkipCron to be true with 'True'")
		}
		if !config.EnableAPI {
			t.Error("Expected EnableAPI to be true with 'True'")
		}
	})
}

func TestLoadConfigMissingRequiredEnvVars(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"DB_PATH", "SMTP_SERVER", "SMTP_PORT", "SMTP_USERNAME",
		"SMTP_PASSWORD", "RECIPIENT_EMAIL", "CHECK_INTERVAL",
		"DEBUG_MOCK_RSS", "DEBUG_SKIP_CRON",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Clean up function to restore environment
	cleanup := func() {
		for _, key := range envVars {
			if val, exists := originalEnv[key]; exists {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}
	defer cleanup()

	testCases := []struct {
		name       string
		missingVar string
		setupEnv   func()
	}{
		{
			name:       "MissingDBPath",
			missingVar: "DB_PATH",
			setupEnv: func() {
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
			},
		},
		{
			name:       "MissingSMTPServer",
			missingVar: "SMTP_SERVER",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
			},
		},
		{
			name:       "MissingSMTPPort",
			missingVar: "SMTP_PORT",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
			},
		},
		{
			name:       "MissingSMTPUsername",
			missingVar: "SMTP_USERNAME",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
			},
		},
		{
			name:       "MissingSMTPPassword",
			missingVar: "SMTP_PASSWORD",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
			},
		},
		{
			name:       "MissingRecipientEmail",
			missingVar: "RECIPIENT_EMAIL",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear environment first
			for _, key := range envVars {
				os.Unsetenv(key)
			}

			// Setup environment without the missing variable
			tc.setupEnv()

			_, err := LoadConfig()
			if err == nil {
				t.Errorf("Expected error when %s is missing, but got none", tc.missingVar)
			}
		})
	}
}

func TestLoadConfigInvalidCheckInterval(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"DB_PATH", "SMTP_SERVER", "SMTP_PORT", "SMTP_USERNAME",
		"SMTP_PASSWORD", "RECIPIENT_EMAIL", "CHECK_INTERVAL",
		"DEBUG_MOCK_RSS", "DEBUG_SKIP_CRON",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Clean up function to restore environment
	cleanup := func() {
		for _, key := range envVars {
			if val, exists := originalEnv[key]; exists {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}
	defer cleanup()

	// Clear environment first
	for _, key := range envVars {
		os.Unsetenv(key)
	}

	// Set required environment variables with invalid CHECK_INTERVAL
	os.Setenv("DB_PATH", "./youtubecurator.db")
	os.Setenv("SMTP_SERVER", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "user@example.com")
	os.Setenv("SMTP_PASSWORD", "password123")
	os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
	os.Setenv("CHECK_INTERVAL", "invalid")

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid CHECK_INTERVAL, but got none")
	}
}
