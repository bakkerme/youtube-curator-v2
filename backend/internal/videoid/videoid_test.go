package videoid

import (
	"testing"
)

func TestNewFromRaw(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name:    "valid raw video ID",
			raw:     "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "valid raw video ID with underscore",
			raw:     "ABC123_def-",
			wantErr: false,
		},
		{
			name:    "invalid raw video ID - too short",
			raw:     "dQw4w9WgX",
			wantErr: true,
		},
		{
			name:    "invalid raw video ID - too long",
			raw:     "dQw4w9WgXcQQ",
			wantErr: true,
		},
		{
			name:    "invalid raw video ID - invalid characters",
			raw:     "dQw4w9WgX@Q",
			wantErr: true,
		},
		{
			name:    "empty raw video ID",
			raw:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vid, err := NewFromRaw(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && vid.ToRaw() != tt.raw {
				t.Errorf("NewFromRaw().ToRaw() = %v, want %v", vid.ToRaw(), tt.raw)
			}
		})
	}
}

func TestNewFromFull(t *testing.T) {
	tests := []struct {
		name    string
		full    string
		wantRaw string
		wantErr bool
	}{
		{
			name:    "valid full video ID",
			full:    "yt:video:dQw4w9WgXcQ",
			wantRaw: "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "valid full video ID with underscore",
			full:    "yt:video:ABC123_def-",
			wantRaw: "ABC123_def-",
			wantErr: false,
		},
		{
			name:    "invalid full video ID - missing prefix",
			full:    "dQw4w9WgXcQ",
			wantErr: true,
		},
		{
			name:    "invalid full video ID - wrong prefix",
			full:    "video:dQw4w9WgXcQ",
			wantErr: true,
		},
		{
			name:    "invalid full video ID - invalid raw part",
			full:    "yt:video:dQw4w9WgX",
			wantErr: true,
		},
		{
			name:    "empty full video ID",
			full:    "",
			wantErr: true,
		},
		{
			name:    "only prefix",
			full:    "yt:video:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vid, err := NewFromFull(tt.full)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromFull() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if vid.ToRaw() != tt.wantRaw {
					t.Errorf("NewFromFull().ToRaw() = %v, want %v", vid.ToRaw(), tt.wantRaw)
				}
				if vid.ToFull() != tt.full {
					t.Errorf("NewFromFull().ToFull() = %v, want %v", vid.ToFull(), tt.full)
				}
			}
		})
	}
}

func TestVideoID_ToFull(t *testing.T) {
	vid, err := NewFromRaw("dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("NewFromRaw() error = %v", err)
	}

	want := "yt:video:dQw4w9WgXcQ"
	if got := vid.ToFull(); got != want {
		t.Errorf("VideoID.ToFull() = %v, want %v", got, want)
	}
}

func TestVideoID_ToRaw(t *testing.T) {
	vid, err := NewFromFull("yt:video:dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("NewFromFull() error = %v", err)
	}

	want := "dQw4w9WgXcQ"
	if got := vid.ToRaw(); got != want {
		t.Errorf("VideoID.ToRaw() = %v, want %v", got, want)
	}
}

func TestVideoID_String(t *testing.T) {
	vid, err := NewFromRaw("dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("NewFromRaw() error = %v", err)
	}

	want := "yt:video:dQw4w9WgXcQ"
	if got := vid.String(); got != want {
		t.Errorf("VideoID.String() = %v, want %v", got, want)
	}
}

func TestValidateFullVideoID(t *testing.T) {
	tests := []struct {
		name    string
		full    string
		wantErr bool
	}{
		{
			name:    "valid full video ID",
			full:    "yt:video:dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "invalid full video ID",
			full:    "dQw4w9WgXcQ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateFullVideoID(tt.full); (err != nil) != tt.wantErr {
				t.Errorf("ValidateFullVideoID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRawVideoID(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name:    "valid raw video ID",
			raw:     "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "invalid raw video ID",
			raw:     "dQw4w9WgX",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRawVideoID(tt.raw); (err != nil) != tt.wantErr {
				t.Errorf("ValidateRawVideoID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
