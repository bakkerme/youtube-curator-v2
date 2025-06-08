package rss

import (
	"strings"
	"testing"
)

func TestValidateYouTubeVideoID(t *testing.T) {
	testCases := []struct {
		name      string
		videoID   string
		expectErr bool
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
