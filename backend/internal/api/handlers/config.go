package handlers

import (
	"net/http"
	"strings"
	"time"

	"youtube-curator-v2/internal/api/types"
	"youtube-curator-v2/internal/store"

	"github.com/labstack/echo/v4"
)

// ConfigHandlers provides handlers for configuration management endpoints
type ConfigHandlers struct {
	*BaseHandlers
}

// NewConfigHandlers creates a new instance of config handlers
func NewConfigHandlers(base *BaseHandlers) *ConfigHandlers {
	return &ConfigHandlers{BaseHandlers: base}
}

// GetCheckInterval handles GET /api/config/interval
func (h *ConfigHandlers) GetCheckInterval(c echo.Context) error {
	interval, err := h.store.GetCheckInterval()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve check interval")
	}

	return c.JSON(http.StatusOK, types.ConfigInterval{Interval: interval.String()})
}

// SetCheckInterval handles PUT /api/config/interval
func (h *ConfigHandlers) SetCheckInterval(c echo.Context) error {
	var req types.ConfigInterval
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Interval == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Interval is required")
	}

	// Parse duration
	duration, err := time.ParseDuration(req.Interval)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid interval format. Use Go duration format (e.g., '1h', '30m', '2h30m')")
	}

	// Validate reasonable range (1 minute to 24 hours)
	if duration < time.Minute {
		return echo.NewHTTPError(http.StatusBadRequest, "Interval must be at least 1 minute")
	}
	if duration > 24*time.Hour {
		return echo.NewHTTPError(http.StatusBadRequest, "Interval must be no more than 24 hours")
	}

	// Set interval in store
	if err := h.store.SetCheckInterval(duration); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set check interval")
	}

	return c.JSON(http.StatusOK, types.ConfigInterval{Interval: duration.String()})
}

// GetSMTPConfig handles GET /api/config/smtp
func (h *ConfigHandlers) GetSMTPConfig(c echo.Context) error {
	config, err := h.store.GetSMTPConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve SMTP configuration")
	}

	// If no config exists, return empty response
	if config == nil {
		return c.JSON(http.StatusOK, types.SMTPConfigResponse{
			PasswordSet: false,
		})
	}

	// Return config without password
	response := types.SMTPConfigResponse{
		Server:         config.Server,
		Port:           config.Port,
		Username:       config.Username,
		RecipientEmail: config.RecipientEmail,
		PasswordSet:    config.Password != "",
	}

	return c.JSON(http.StatusOK, response)
}

// SetSMTPConfig handles PUT /api/config/smtp
func (h *ConfigHandlers) SetSMTPConfig(c echo.Context) error {
	var req types.SMTPConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Server == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server is required")
	}
	if req.Port == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Port is required")
	}
	if req.Username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Username is required")
	}
	if req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Password is required")
	}
	if req.RecipientEmail == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Recipient email is required")
	}

	// Basic email validation
	if !strings.Contains(req.RecipientEmail, "@") {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid recipient email format")
	}

	// Create SMTP config
	smtpConfig := &store.SMTPConfig{
		Server:         req.Server,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		RecipientEmail: req.RecipientEmail,
	}

	// Save to store
	if err := h.store.SetSMTPConfig(smtpConfig); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save SMTP configuration")
	}

	// Return response without password
	response := types.SMTPConfigResponse{
		Server:         req.Server,
		Port:           req.Port,
		Username:       req.Username,
		RecipientEmail: req.RecipientEmail,
		PasswordSet:    true,
	}

	return c.JSON(http.StatusOK, response)
}

// GetLLMConfig handles GET /api/config/llm
func (h *ConfigHandlers) GetLLMConfig(c echo.Context) error {
	config, err := h.store.GetLLMConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve LLM configuration")
	}

	// If no config exists, return empty response
	if config == nil {
		return c.JSON(http.StatusOK, types.LLMConfigResponse{
			APIKeySet: false,
		})
	}

	// Return config without API key
	response := types.LLMConfigResponse{
		EndpointURL: config.EndpointURL,
		Model:       config.Model,
		APIKeySet:   config.APIKey != "",
	}

	return c.JSON(http.StatusOK, response)
}

// SetLLMConfig handles PUT /api/config/llm
func (h *ConfigHandlers) SetLLMConfig(c echo.Context) error {
	var req types.LLMConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.EndpointURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Endpoint URL is required")
	}
	if req.APIKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "API key is required")
	}
	if req.Model == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Model is required")
	}

	// Create LLM config
	llmConfig := &store.LLMConfig{
		EndpointURL: req.EndpointURL,
		APIKey:      req.APIKey,
		Model:       req.Model,
	}

	// Save to store
	if err := h.store.SetLLMConfig(llmConfig); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save LLM configuration")
	}

	// Return response without API key
	response := types.LLMConfigResponse{
		EndpointURL: req.EndpointURL,
		Model:       req.Model,
		APIKeySet:   true,
	}

	return c.JSON(http.StatusOK, response)
}