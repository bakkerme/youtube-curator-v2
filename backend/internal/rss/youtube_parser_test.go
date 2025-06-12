package rss

import (
	"context"
	"errors"
	"testing"
)

// MockResolver is a mock implementation of ChannelIDResolver for testing
type MockResolver struct {
	ShouldFail bool
	ReturnID   string
}

// ResolveChannelID implements the ChannelIDResolver interface
func (m *MockResolver) ResolveChannelID(ctx context.Context, url string) (string, error) {
	if m.ShouldFail {
		return "", errors.New("mock resolver failed")
	}

	if m.ReturnID != "" {
		return m.ReturnID, nil
	}

	// Default successful resolution - return a valid channel ID
	return "UCrAhw9Z8NI6GzO2WnvhYzCg", nil
}

func TestValidateYouTubeVideoID_Refactored(t *testing.T) {
	testCases := []struct {
		name      string
		videoID   string
		expectErr bool
		errType   error // Check for specific error type instead of string content
	}{
		// Valid cases
		{name: "Valid ID standard", videoID: "yt:video:dQw4w9WgXcQ", expectErr: false},
		{name: "Valid ID with hyphens and underscores", videoID: "yt:video:a_b-c_12345", expectErr: false},
		{name: "Valid ID all numbers", videoID: "yt:video:12345678901", expectErr: false},
		{name: "Valid ID all lowercase", videoID: "yt:video:abcdefghijk", expectErr: false},
		{name: "Valid ID all uppercase", videoID: "yt:video:ABCDEFGHIJK", expectErr: false},

		// Invalid cases - all should return NewInvalidVideoIDError()
		{name: "Invalid prefix - wrong start", videoID: "youtube:video:dQw4w9WgXcQ", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "Invalid prefix - incomplete", videoID: "yt:vid:dQw4w9WgXcQ", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "Invalid prefix - no prefix, just ID", videoID: "dQw4w9WgXcQ", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "Incorrect length - too short", videoID: "yt:video:short", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "Incorrect length - too long", videoID: "yt:video:toolong123456", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "Invalid characters - symbols", videoID: "yt:video:!@#$%^&*()_+", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "Invalid characters - space", videoID: "yt:video:valid space", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "Empty string", videoID: "", expectErr: true, errType: NewInvalidVideoIDError()},
		{name: "String is just the prefix", videoID: "yt:video:", expectErr: true, errType: NewInvalidVideoIDError()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateYouTubeVideoID(tc.videoID)
			if tc.expectErr {
				if err == nil {
					t.Errorf("ValidateYouTubeVideoID(%q) expected error, but got nil", tc.videoID)
				} else {
					// Check if the error is of the expected type by comparing the underlying ValidationError
					var validationErr ValidationError
					if errors.As(err, &validationErr) {
						var expectedValidationErr ValidationError
						if errors.As(tc.errType, &expectedValidationErr) {
							if validationErr.Type != expectedValidationErr.Type || validationErr.Field != expectedValidationErr.Field {
								t.Errorf("ValidateYouTubeVideoID(%q) expected error type %+v, but got %+v", tc.videoID, expectedValidationErr, validationErr)
							}
						} else {
							t.Errorf("ValidateYouTubeVideoID(%q) expected ValidationError, but errType is not ValidationError", tc.videoID)
						}
					} else {
						t.Errorf("ValidateYouTubeVideoID(%q) expected ValidationError, but got %T: %v", tc.videoID, err, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ValidateYouTubeVideoID(%q) expected no error, but got: %v", tc.videoID, err)
				}
			}
		})
	}
}

func TestExtractChannelIDWithResolver_Refactored(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		resolver   ChannelIDResolver
		expectedID string
		expectErr  bool
		errType    error // Check for specific error type instead of string content
	}{
		// Test cases that should work without resolver (backwards compatibility)
		{
			name:       "Valid channel ID - no resolver needed",
			input:      "UCrAhw9Z8NI6GzO2WnvhYzCg",
			resolver:   nil,
			expectedID: "UCrAhw9Z8NI6GzO2WnvhYzCg",
			expectErr:  false,
		},
		{
			name:       "Direct channel URL - no resolver needed",
			input:      "https://www.youtube.com/channel/UCrAhw9Z8NI6GzO2WnvhYzCg",
			resolver:   nil,
			expectedID: "UCrAhw9Z8NI6GzO2WnvhYzCg",
			expectErr:  false,
		},

		// Test cases that require resolver - using specific error types
		{
			name:      "@username URL without resolver",
			input:     "https://youtube.com/@ChinaTalkMedi",
			resolver:  nil,
			expectErr: true,
			errType:   NewResolverRequiredError("@username"),
		},
		{
			name:      "/c/ URL without resolver",
			input:     "https://youtube.com/c/somechannel",
			resolver:  nil,
			expectErr: true,
			errType:   NewResolverRequiredError("custom URLs (/c/, /user/)"),
		},
		{
			name:      "/user/ URL without resolver",
			input:     "https://youtube.com/user/someuser",
			resolver:  nil,
			expectErr: true,
			errType:   NewResolverRequiredError("custom URLs (/c/, /user/)"),
		},

		// Test cases with working resolver
		{
			name:       "@username URL with resolver",
			input:      "https://youtube.com/@ChinaTalkMedi",
			resolver:   &MockResolver{},
			expectedID: "UCrAhw9Z8NI6GzO2WnvhYzCg",
			expectErr:  false,
		},

		// Test error cases with specific error types
		{
			name:      "Resolver fails",
			input:     "https://youtube.com/@test",
			resolver:  &MockResolver{ShouldFail: true},
			expectErr: true,
			errType:   ErrResolverFailed,
		},
		{
			name:      "Resolver returns invalid channel ID",
			input:     "https://youtube.com/@test",
			resolver:  &MockResolver{ReturnID: "invalid-id"},
			expectErr: true,
			errType:   ErrResolvedIDInvalid,
		},
		{
			name:      "Invalid URL format",
			input:     "not a url",
			resolver:  nil,
			expectErr: true,
			errType:   ErrUnsupportedURLFormat,
		},
		{
			name:      "Unsupported URL format",
			input:     "https://youtube.com/unknown/path",
			resolver:  nil,
			expectErr: true,
			errType:   ErrUnsupportedURLFormat,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channelID, err := ExtractChannelIDWithResolver(ctx, tc.input, tc.resolver)

			if tc.expectErr {
				if err == nil {
					t.Errorf("ExtractChannelIDWithResolver(%q) expected error, but got nil", tc.input)
				} else {
					// Check error type using errors.Is for standard errors
					if tc.errType != nil {
						var validationErr ValidationError
						if errors.As(tc.errType, &validationErr) {
							// For ValidationError, compare type and field
							var actualValidationErr ValidationError
							if errors.As(err, &actualValidationErr) {
								if actualValidationErr.Type != validationErr.Type {
									t.Errorf("ExtractChannelIDWithResolver(%q) expected ValidationError type %q, but got %q", tc.input, validationErr.Type, actualValidationErr.Type)
								}
							} else {
								t.Errorf("ExtractChannelIDWithResolver(%q) expected ValidationError, but got %T: %v", tc.input, err, err)
							}
						} else {
							// For standard errors, use errors.Is
							if !errors.Is(err, tc.errType) {
								t.Errorf("ExtractChannelIDWithResolver(%q) expected error %v, but got %v", tc.input, tc.errType, err)
							}
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("ExtractChannelIDWithResolver(%q) expected no error, but got: %v", tc.input, err)
				} else if channelID != tc.expectedID {
					t.Errorf("ExtractChannelIDWithResolver(%q) expected channel ID %q, but got %q", tc.input, tc.expectedID, channelID)
				}
			}
		})
	}
}

// TestValidateChannelID_Refactored demonstrates robust error testing
func TestValidateChannelID_Refactored(t *testing.T) {
	testCases := []struct {
		name      string
		channelID string
		expectErr bool
		errType   error
	}{
		{name: "Valid channel ID", channelID: "UCrAhw9Z8NI6GzO2WnvhYzCg", expectErr: false},
		{name: "Invalid channel ID - too short", channelID: "UC123", expectErr: true, errType: NewInvalidChannelIDError("UC123")},
		{name: "Invalid channel ID - wrong prefix", channelID: "XC1234567890123456789012", expectErr: true, errType: NewInvalidChannelIDError("XC1234567890123456789012")},
		{name: "Invalid channel ID - too long", channelID: "UCrAhw9Z8NI6GzO2WnvhYzCg123", expectErr: true, errType: NewInvalidChannelIDError("UCrAhw9Z8NI6GzO2WnvhYzCg123")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateChannelID(tc.channelID)
			if tc.expectErr {
				if err == nil {
					t.Errorf("ValidateChannelID(%q) expected error, but got nil", tc.channelID)
				} else {
					// Check if the error is of the expected ValidationError type
					var validationErr ValidationError
					if errors.As(err, &validationErr) {
						var expectedValidationErr ValidationError
						if errors.As(tc.errType, &expectedValidationErr) {
							if validationErr.Type != expectedValidationErr.Type || validationErr.Field != expectedValidationErr.Field {
								t.Errorf("ValidateChannelID(%q) expected error type %+v, but got %+v", tc.channelID, expectedValidationErr, validationErr)
							}
						} else {
							t.Errorf("ValidateChannelID(%q) expected ValidationError, but errType is not ValidationError", tc.channelID)
						}
					} else {
						t.Errorf("ValidateChannelID(%q) expected ValidationError, but got %T: %v", tc.channelID, err, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ValidateChannelID(%q) expected no error, but got: %v", tc.channelID, err)
				}
			}
		})
	}
}
