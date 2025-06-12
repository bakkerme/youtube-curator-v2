package types

// ChannelRequest represents a request to add a channel
// Title is optional; if not provided, it will be fetched from RSS
type ChannelRequest struct {
	URL   string `json:"url" validate:"required"`
	Title string `json:"title,omitempty"`
}

// ConfigInterval represents the check interval configuration
type ConfigInterval struct {
	Interval string `json:"interval"`
}

// SMTPConfigRequest represents a request to update SMTP configuration
type SMTPConfigRequest struct {
	Server         string `json:"server" validate:"required"`
	Port           string `json:"port" validate:"required"`
	Username       string `json:"username" validate:"required"`
	Password       string `json:"password" validate:"required"`
	RecipientEmail string `json:"recipientEmail" validate:"required,email"`
}

// LLMConfigRequest represents a request to update LLM configuration
type LLMConfigRequest struct {
	EndpointURL string `json:"endpoint" validate:"required"`
	APIKey      string `json:"apiKey" validate:"required"`
	Model       string `json:"model" validate:"required"`
}

// ImportChannelsRequest represents a request to import multiple channels
type ImportChannelsRequest struct {
	Channels []ChannelImport `json:"channels" validate:"required"`
}

// ChannelImport represents a channel to be imported
type ChannelImport struct {
	URL   string `json:"url" validate:"required"`
	Title string `json:"title,omitempty"`
}

// RunNewsletterRequest represents a request to manually trigger newsletter run
type RunNewsletterRequest struct {
	ChannelID         string `json:"channelId,omitempty"`
	IgnoreLastChecked bool   `json:"ignoreLastChecked,omitempty"`
	MaxItems          int    `json:"maxItems,omitempty"`
}