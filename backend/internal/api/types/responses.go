package types

import "time"

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage  int  `json:"currentPage"`
	TotalPages   int  `json:"totalPages"`
	TotalItems   int  `json:"totalItems"`
	ItemsPerPage int  `json:"itemsPerPage"`
	HasNext      bool `json:"hasNext"`
	HasPrevious  bool `json:"hasPrevious"`
}

// ChannelResponse represents a channel in API responses
type ChannelResponse struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	CustomURL            string    `json:"customUrl,omitempty"`
	ThumbnailURL         string    `json:"thumbnailUrl,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	LastVideoPublishedAt time.Time `json:"lastVideoPublishedAt,omitempty"`
	VideoCount           int       `json:"videoCount"`
	IsActive             bool      `json:"isActive"`
}

// ChannelsResponse represents the response for GET /api/channels
type ChannelsResponse struct {
	Channels    []ChannelResponse `json:"channels"`
	TotalCount  int               `json:"totalCount"`
	LastUpdated time.Time         `json:"lastUpdated"`
}

// VideoLinkResponse represents a video link in API responses
type VideoLinkResponse struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// VideoAuthorResponse represents a video author in API responses
type VideoAuthorResponse struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// VideoMediaThumbnailResponse represents a video thumbnail in API responses
type VideoMediaThumbnailResponse struct {
	URL    string `json:"url"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

// VideoMediaContentResponse represents video media content in API responses
type VideoMediaContentResponse struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

// VideoMediaGroupResponse represents video media group in API responses
type VideoMediaGroupResponse struct {
	MediaThumbnail   VideoMediaThumbnailResponse `json:"mediaThumbnail"`
	MediaTitle       string                      `json:"mediaTitle"`
	MediaContent     VideoMediaContentResponse   `json:"mediaContent"`
	MediaDescription string                      `json:"mediaDescription"`
}

// VideoResponse represents a video in API responses matching the frontend VideoEntry interface
type VideoResponse struct {
	ID         string                  `json:"id"`
	ChannelID  string                  `json:"channelId"`
	CachedAt   time.Time               `json:"cachedAt"`
	Watched    bool                    `json:"watched"`
	Title      string                  `json:"title"`
	Link       VideoLinkResponse       `json:"link"`
	Published  time.Time               `json:"published"`
	Content    string                  `json:"content"`
	Author     VideoAuthorResponse     `json:"author"`
	MediaGroup VideoMediaGroupResponse `json:"mediaGroup"`
	Summary    *VideoSummaryInfo       `json:"summary,omitempty"` // Video summary if available
}

// VideoSummaryInfo represents summary information in video responses
type VideoSummaryInfo struct {
	Text               string `json:"text"`
	SourceLanguage     string `json:"sourceLanguage"`
	SummaryGeneratedAt string `json:"summaryGeneratedAt"` // ISO 8601 format
}

// VideosResponse represents the response for GET /api/videos
type VideosResponse struct {
	Videos      []VideoResponse `json:"videos"`
	TotalCount  int             `json:"totalCount"`
	LastRefresh time.Time       `json:"lastRefresh"`
	Pagination  *Pagination     `json:"pagination,omitempty"`
}

// NewsletterRunResponse represents the response from triggering a newsletter run
type NewsletterRunResponse struct {
	Message           string `json:"message"`
	ChannelsProcessed int    `json:"channelsProcessed"`
	ChannelsWithError int    `json:"channelsWithError"`
	NewVideosFound    int    `json:"newVideosFound"`
	EmailSent         bool   `json:"emailSent"`
}

// SMTPConfigResponse represents SMTP configuration in API responses (without password)
type SMTPConfigResponse struct {
	Server         string `json:"server"`
	Port           string `json:"port"`
	Username       string `json:"username"`
	RecipientEmail string `json:"recipientEmail"`
	PasswordSet    bool   `json:"passwordSet"`
}

// LLMConfigResponse represents LLM configuration in API responses (without API key)
type LLMConfigResponse struct {
	EndpointURL string `json:"endpointUrl"`
	Model       string `json:"model"`
	APIKeySet   bool   `json:"apiKeySet"`
}

// NewsletterConfigResponse represents newsletter configuration in API responses
type NewsletterConfigResponse struct {
	Enabled bool `json:"enabled"`
}

// VideoSummaryResponse represents a video summary in API responses
type VideoSummaryResponse struct {
	VideoID        string `json:"videoId"`
	Summary        string `json:"summary"`
	Thinking       string `json:"thinking,omitempty"`  // LLM thinking content from <think> blocks
	SourceLanguage string `json:"sourceLanguage"`
	GeneratedAt    string `json:"generatedAt"` // ISO 8601 format
	Tracked        bool   `json:"tracked"`     // Whether this video is from a tracked channel
}

// ImportChannelsResponse represents the response from importing channels
type ImportChannelsResponse struct {
	Imported []ChannelResponse `json:"imported"`
	Failed   []ImportFailure   `json:"failed"`
}

// ImportFailure represents a failed channel import
type ImportFailure struct {
	Channel ChannelImport `json:"channel"`
	Error   string        `json:"error"`
}

// Channel represents a channel in API responses (legacy, use ChannelResponse instead)
type Channel struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}