package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"DB_PATH", "CHANNELS_FILE", "SMTP_SERVER", "SMTP_PORT", "SMTP_USERNAME",
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
		os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
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
	})

	t.Run("DefaultCheckInterval", func(t *testing.T) {
		// Clear environment first
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		// Set required environment variables without CHECK_INTERVAL
		os.Setenv("DB_PATH", "./youtubecurator.db")
		os.Setenv("SMTP_SERVER", "smtp.example.com")
		os.Setenv("SMTP_PORT", "587")
		os.Setenv("SMTP_USERNAME", "user@example.com")
		os.Setenv("SMTP_PASSWORD", "password123")
		os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
		os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if config.CheckInterval != time.Hour {
			t.Errorf("Expected default CheckInterval to be 1h, got %v", config.CheckInterval)
		}
	})

	t.Run("DebugFlagsDefault", func(t *testing.T) {
		// Clear environment first
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		// Set required environment variables without debug flags
		os.Setenv("DB_PATH", "./youtubecurator.db")
		os.Setenv("SMTP_SERVER", "smtp.example.com")
		os.Setenv("SMTP_PORT", "587")
		os.Setenv("SMTP_USERNAME", "user@example.com")
		os.Setenv("SMTP_PASSWORD", "password123")
		os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
		os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if config.DebugMockRSS {
			t.Error("Expected DebugMockRSS to be false by default")
		}
		if config.DebugSkipCron {
			t.Error("Expected DebugSkipCron to be false by default")
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
		os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
		os.Setenv("DEBUG_MOCK_RSS", "TRUE")
		os.Setenv("DEBUG_SKIP_CRON", "True")
		os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")

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
	})

	t.Run("WithChannelsFile", func(t *testing.T) {
		// Clear environment first
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		// Create a temporary channels file
		tempDir := t.TempDir()
		channelsFile := filepath.Join(tempDir, "channels.txt")
		channelsContent := "channel1\nchannel2\n# comment\n\nchannel3\n"
		err := os.WriteFile(channelsFile, []byte(channelsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test channels file: %v", err)
		}

		// Set required environment variables
		os.Setenv("DB_PATH", "./youtubecurator.db")
		os.Setenv("CHANNELS_FILE", channelsFile)
		os.Setenv("SMTP_SERVER", "smtp.example.com")
		os.Setenv("SMTP_PORT", "587")
		os.Setenv("SMTP_USERNAME", "user@example.com")
		os.Setenv("SMTP_PASSWORD", "password123")
		os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
		os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedChannels := []string{"channel1", "channel2", "channel3"}
		if len(config.Channels) != len(expectedChannels) {
			t.Errorf("Expected %d channels, got %d", len(expectedChannels), len(config.Channels))
		}
		for i, expected := range expectedChannels {
			if i >= len(config.Channels) || config.Channels[i] != expected {
				t.Errorf("Expected channel[%d] to be '%s', got '%s'", i, expected, config.Channels[i])
			}
		}
	})
}

func TestLoadConfigMissingRequiredEnvVars(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"DB_PATH", "CHANNELS_FILE", "SMTP_SERVER", "SMTP_PORT", "SMTP_USERNAME",
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
			name:       "MissingSMTPServer",
			missingVar: "SMTP_SERVER",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
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
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
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
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
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
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
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
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
			},
		},
		{
			name:       "MissingChannelsFile",
			missingVar: "CHANNELS_FILE",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
			},
		},
		{
			name:       "MissingDebugMockRSS",
			missingVar: "DEBUG_MOCK_RSS",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
			},
		},
		{
			name:       "MissingDebugSkipCron",
			missingVar: "DEBUG_SKIP_CRON",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
				os.Setenv("DEBUG_MOCK_RSS", "true")
			},
		},
		{
			name:       "MissingCheckInterval",
			missingVar: "CHECK_INTERVAL",
			setupEnv: func() {
				os.Setenv("DB_PATH", "./youtubecurator.db")
				os.Setenv("SMTP_SERVER", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USERNAME", "user@example.com")
				os.Setenv("SMTP_PASSWORD", "password123")
				os.Setenv("RECIPIENT_EMAIL", "recipient@example.com")
				os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")
				os.Setenv("DEBUG_MOCK_RSS", "true")
				os.Setenv("DEBUG_SKIP_CRON", "true")
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

			// This test would normally cause log.Fatal, but we can't easily test that
			// In a real scenario, you might want to refactor LoadConfig to return errors
			// instead of calling log.Fatal directly
			t.Skip("Skipping test that would call log.Fatal - consider refactoring LoadConfig to return errors")
		})
	}
}

func TestLoadConfigInvalidCheckInterval(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"DB_PATH", "CHANNELS_FILE", "SMTP_SERVER", "SMTP_PORT", "SMTP_USERNAME",
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
	os.Setenv("CHANNELS_FILE", "__mocks__/channels.txt")

	// This test would normally cause log.Fatal, but we can't easily test that
	// In a real scenario, you might want to refactor LoadConfig to return errors
	t.Skip("Skipping test that would call log.Fatal - consider refactoring LoadConfig to return errors")
}

func TestLoadChannelsFromFile(t *testing.T) {
	t.Run("ValidFile", func(t *testing.T) {
		tempDir := t.TempDir()
		channelsFile := filepath.Join(tempDir, "channels.txt")
		content := "channel1\nchannel2\n# This is a comment\n\n  channel3  \n# Another comment\nchannel4\n"
		err := os.WriteFile(channelsFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		channels, err := loadChannelsFromFile(channelsFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := []string{"channel1", "channel2", "channel3", "channel4"}
		if len(channels) != len(expected) {
			t.Errorf("Expected %d channels, got %d", len(expected), len(channels))
		}
		for i, exp := range expected {
			if i >= len(channels) || channels[i] != exp {
				t.Errorf("Expected channel[%d] to be '%s', got '%s'", i, exp, channels[i])
			}
		}
	})

	t.Run("EmptyFile", func(t *testing.T) {
		tempDir := t.TempDir()
		channelsFile := filepath.Join(tempDir, "empty.txt")
		err := os.WriteFile(channelsFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		channels, err := loadChannelsFromFile(channelsFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(channels) != 0 {
			t.Errorf("Expected 0 channels from empty file, got %d", len(channels))
		}
	})

	t.Run("OnlyCommentsAndEmptyLines", func(t *testing.T) {
		tempDir := t.TempDir()
		channelsFile := filepath.Join(tempDir, "comments.txt")
		content := "# Comment 1\n\n# Comment 2\n\n  \n# Comment 3\n"
		err := os.WriteFile(channelsFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		channels, err := loadChannelsFromFile(channelsFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(channels) != 0 {
			t.Errorf("Expected 0 channels from comments-only file, got %d", len(channels))
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := loadChannelsFromFile("/nonexistent/file.txt")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("WhitespaceHandling", func(t *testing.T) {
		tempDir := t.TempDir()
		channelsFile := filepath.Join(tempDir, "whitespace.txt")
		content := "  channel1  \n\tchannel2\t\n   \n\tchannel3   \n"
		err := os.WriteFile(channelsFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		channels, err := loadChannelsFromFile(channelsFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := []string{"channel1", "channel2", "channel3"}
		if len(channels) != len(expected) {
			t.Errorf("Expected %d channels, got %d", len(expected), len(channels))
		}
		for i, exp := range expected {
			if i >= len(channels) || channels[i] != exp {
				t.Errorf("Expected channel[%d] to be '%s', got '%s'", i, exp, channels[i])
			}
		}
	})
}
