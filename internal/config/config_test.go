package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("CHANNELS", "channel1, channel2")
	os.Setenv("SMTP_SERVER", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "user")
	os.Setenv("SMTP_PASSWORD", "pass")
	os.Setenv("RECIPIENT_EMAIL", "test@example.com")
	os.Setenv("CHECK_INTERVAL", "30m")

	cfg := LoadConfig()

	if len(cfg.Channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(cfg.Channels))
	}
	if cfg.Channels[0] != "channel1" || cfg.Channels[1] != "channel2" {
		t.Errorf("Unexpected channel values: %v", cfg.Channels)
	}
	if cfg.SMTPServer != "smtp.example.com" {
		t.Errorf("Expected SMTP_SERVER to be smtp.example.com, got %s", cfg.SMTPServer)
	}
	if cfg.SMTPPort != "587" {
		t.Errorf("Expected SMTP_PORT to be 587, got %s", cfg.SMTPPort)
	}
	if cfg.SMTPUsername != "user" {
		t.Errorf("Expected SMTP_USERNAME to be user, got %s", cfg.SMTPUsername)
	}
	if cfg.SMTPPassword != "pass" {
		t.Errorf("Expected SMTP_PASSWORD to be pass, got %s", cfg.SMTPPassword)
	}
	if cfg.RecipientEmail != "test@example.com" {
		t.Errorf("Expected RECIPIENT_EMAIL to be test@example.com, got %s", cfg.RecipientEmail)
	}
	if cfg.CheckInterval != 30*time.Minute {
		t.Errorf("Expected CHECK_INTERVAL to be 30m, got %s", cfg.CheckInterval)
	}

	// Clean up environment variables
	os.Unsetenv("CHANNELS")
	os.Unsetenv("SMTP_SERVER")
	os.Unsetenv("SMTP_PORT")
	os.Unsetenv("SMTP_USERNAME")
	os.Unsetenv("SMTP_PASSWORD")
	os.Unsetenv("RECIPIENT_EMAIL")
	os.Unsetenv("CHECK_INTERVAL")
}

func TestLoadConfigDefaults(t *testing.T) {
	// Set up minimal required environment variables
	os.Setenv("CHANNELS", "channel1")
	os.Setenv("SMTP_SERVER", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "user")
	os.Setenv("SMTP_PASSWORD", "pass")
	os.Setenv("RECIPIENT_EMAIL", "test@example.com")

	// CHECK_INTERVAL is not set, should default
	cfg := LoadConfig()

	if cfg.CheckInterval != 1*time.Hour {
		t.Errorf("Expected default CHECK_INTERVAL to be 1h, got %s", cfg.CheckInterval)
	}

	// Clean up environment variables
	os.Unsetenv("CHANNELS")
	os.Unsetenv("SMTP_SERVER")
	os.Unsetenv("SMTP_PORT")
	os.Unsetenv("SMTP_USERNAME")
	os.Unsetenv("SMTP_PASSWORD")
	os.Unsetenv("RECIPIENT_EMAIL")
}
