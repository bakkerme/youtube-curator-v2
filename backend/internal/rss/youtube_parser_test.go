package rss

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestValidateYouTubeVideoID(t *testing.T) {
	testCases := []struct {
		name        string
		videoID     string
		expectErr   bool
		errContains string // Optional: check if error message contains this substring
	}{
		// Valid cases
		{name: "Valid ID standard", videoID: "yt:video:dQw4w9WgXcQ", expectErr: false},
		{name: "Valid ID with hyphens and underscores", videoID: "yt:video:a_b-c_12345", expectErr: false},
		{name: "Valid ID all numbers", videoID: "yt:video:12345678901", expectErr: false},
		{name: "Valid ID all lowercase", videoID: "yt:video:abcdefghijk", expectErr: false},
		{name: "Valid ID all uppercase", videoID: "yt:video:ABCDEFGHIJK", expectErr: false},

		// Invalid prefix
		{name: "Invalid prefix - wrong start", videoID: "youtube:video:dQw4w9WgXcQ", expectErr: true, errContains: "Expected prefix 'yt:video:'"},
		{name: "Invalid prefix - incomplete", videoID: "yt:vid:dQw4w9WgXcQ", expectErr: true, errContains: "Expected prefix 'yt:video:'"},
		{name: "Invalid prefix - no prefix, just ID", videoID: "dQw4w9WgXcQ", expectErr: true, errContains: "Expected prefix 'yt:video:'"},

		// Incorrect length for the ID part
		{name: "Incorrect length - too short", videoID: "yt:video:short", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},
		{name: "Incorrect length - too long", videoID: "yt:video:toolong123456", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},

		// Invalid characters in the ID part
		{name: "Invalid characters - symbols", videoID: "yt:video:!@#$%^&*()_+", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},
		{name: "Invalid characters - space", videoID: "yt:video:valid space", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},
		{name: "Invalid characters - trailing dot", videoID: "yt:video:validButEnd.", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},
		{name: "Invalid characters - leading dot", videoID: "yt:video:.validStart", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},

		// Edge cases
		{name: "Empty string", videoID: "", expectErr: true, errContains: "Expected prefix 'yt:video:'"},
		{name: "String is just the prefix", videoID: "yt:video:", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},
		{name: "String is prefix and one char", videoID: "yt:video:a", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},
		{name: "String is prefix and 10 chars", videoID: "yt:video:abcdefghij", expectErr: true, errContains: "Expected format: yt:video:<11_alphanumeric_chars_hyphens_underscores>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateYouTubeVideoID(tc.videoID)
			if tc.expectErr {
				if err == nil {
					t.Errorf("ValidateYouTubeVideoID(%q) expected error, but got nil", tc.videoID)
				} else if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("ValidateYouTubeVideoID(%q) expected error message to contain %q, but got %q", tc.videoID, tc.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("ValidateYouTubeVideoID(%q) expected no error, but got: %v", tc.videoID, err)
				}
			}
		})
	}
}

// MockResolver is a test implementation of ChannelIDResolver
type MockResolver struct {
	ShouldFail bool
	ReturnID   string
}

func (m *MockResolver) ResolveChannelID(ctx context.Context, url string) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock resolver failed")
	}
	if m.ReturnID != "" {
		return m.ReturnID, nil
	}
	// Default behavior: return a mock channel ID based on URL
	if strings.Contains(url, "@ChinaTalkMedi") {
		return "UCrAhw9Z8NI6GzO2WnvhYzCg", nil
	}
	return "UCMockChannelID1234567890", nil
}

func TestExtractChannelIDWithResolver(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		resolver    ChannelIDResolver
		expectedID  string
		expectErr   bool
		errContains string
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

		// Test cases that require resolver
		{
			name:        "@username URL without resolver",
			input:       "https://youtube.com/@ChinaTalkMedi",
			resolver:    nil,
			expectErr:   true,
			errContains: "require a resolver",
		},
		{
			name:        "/c/ URL without resolver",
			input:       "https://youtube.com/c/somechannel",
			resolver:    nil,
			expectErr:   true,
			errContains: "require a resolver",
		},
		{
			name:        "/user/ URL without resolver",
			input:       "https://youtube.com/user/someuser",
			resolver:    nil,
			expectErr:   true,
			errContains: "require a resolver",
		},

		// Test cases with working resolver
		{
			name:       "@username URL with resolver",
			input:      "https://youtube.com/@ChinaTalkMedi",
			resolver:   &MockResolver{},
			expectedID: "UCrAhw9Z8NI6GzO2WnvhYzCg",
			expectErr:  false,
		},
		{
			name:       "/c/ URL with resolver",
			input:      "https://youtube.com/c/somechannel",
			resolver:   &MockResolver{ReturnID: "UCTestChannel123456789ab"},
			expectedID: "UCTestChannel123456789ab",
			expectErr:  false,
		},
		{
			name:       "/user/ URL with resolver",
			input:      "https://youtube.com/user/someuser",
			resolver:   &MockResolver{ReturnID: "UCUserChannel123456789cd"},
			expectedID: "UCUserChannel123456789cd",
			expectErr:  false,
		},

		// Test error cases
		{
			name:        "Resolver fails",
			input:       "https://youtube.com/@test",
			resolver:    &MockResolver{ShouldFail: true},
			expectErr:   true,
			errContains: "failed to resolve channel ID",
		},
		{
			name:        "Resolver returns invalid channel ID",
			input:       "https://youtube.com/@test",
			resolver:    &MockResolver{ReturnID: "invalid-id"},
			expectErr:   true,
			errContains: "not in valid format",
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channelID, err := ExtractChannelIDWithResolver(ctx, tc.input, tc.resolver)

			if tc.expectErr {
				if err == nil {
					t.Errorf("ExtractChannelIDWithResolver(%q) expected error, but got nil", tc.input)
				} else if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("ExtractChannelIDWithResolver(%q) expected error message to contain %q, but got %q", tc.input, tc.errContains, err.Error())
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
