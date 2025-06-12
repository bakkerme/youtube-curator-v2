package rss

import "errors"

// Define error types for better testability and consistency
var (
	// URL/Input validation errors
	ErrInvalidURL           = errors.New("invalid URL or channel ID")
	ErrInvalidURLFormat     = errors.New("invalid YouTube URL format")
	ErrUnsupportedURLFormat = errors.New("unsupported YouTube URL format")

	// Channel ID validation errors
	ErrInvalidChannelIDFormat = errors.New("invalid channel ID format")
	ErrInvalidVideoIDFormat   = errors.New("invalid video ID format")

	// Resolver requirement errors
	ErrResolverRequired  = errors.New("resolver required for this URL type")
	ErrResolverFailed    = errors.New("failed to resolve channel ID")
	ErrResolvedIDInvalid = errors.New("resolved channel ID is not in valid format")
)

// Error types for different categories of validation failures
type ValidationError struct {
	Type    string
	Field   string
	Value   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// Helper functions to create specific validation errors
func NewInvalidChannelIDError(channelID string) ValidationError {
	return ValidationError{
		Type:    "validation",
		Field:   "channel_id",
		Value:   channelID,
		Message: "invalid channel ID format. Channel IDs should start with 'UC' and be 24 characters long",
	}
}

func NewInvalidVideoIDError() ValidationError {
	return ValidationError{
		Type:    "validation",
		Field:   "video_id",
		Value:   "",
		Message: "invalid video ID format. Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>",
	}
}

func NewResolverRequiredError(urlType string) ValidationError {
	return ValidationError{
		Type:    "resolver_required",
		Field:   "url",
		Value:   urlType,
		Message: urlType + " URLs require a resolver. Please provide a ChannelIDResolver or use the channel ID directly (starts with 'UC')",
	}
}
