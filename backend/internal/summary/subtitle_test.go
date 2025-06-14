package summary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseSubtitleContent(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "VTT format",
			input: `WEBVTT

00:00:01.000 --> 00:00:04.000
Hello and welcome to this video.

00:00:04.000 --> 00:00:08.000
Today we're going to learn about Go programming.

00:00:08.000 --> 00:00:12.000
Let's start with the basics.`,
			expected: "Hello and welcome to this video. Today we're going to learn about Go programming. Let's start with the basics.",
		},
		{
			name: "SRT format",
			input: `1
00:00:01,000 --> 00:00:04,000
Hello and welcome to this video.

2
00:00:04,000 --> 00:00:08,000
Today we're going to learn about Go programming.

3
00:00:08,000 --> 00:00:12,000
Let's start with the basics.`,
			expected: "Hello and welcome to this video. Today we're going to learn about Go programming. Let's start with the basics.",
		},
		{
			name: "With HTML tags",
			input: `WEBVTT

00:00:01.000 --> 00:00:04.000
<c>Hello</c> and <b>welcome</b> to this video.

00:00:04.000 --> 00:00:08.000
Today we&apos;re going to learn about &quot;Go&quot; programming.`,
			expected: `Hello and welcome to this video. Today we're going to learn about "Go" programming.`,
		},
		{
			name: "Empty content",
			input: `WEBVTT

00:00:01.000 --> 00:00:04.000


00:00:04.000 --> 00:00:08.000
   

00:00:08.000 --> 00:00:12.000
Valid content here.`,
			expected: "Valid content here.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.parseSubtitleContent(tt.input)
			if result != tt.expected {
				t.Errorf("parseSubtitleContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFetchSubtitleText(t *testing.T) {
	// Create a test server that serves subtitle content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/vtt")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`WEBVTT

00:00:01.000 --> 00:00:04.000
This is a test subtitle.

00:00:04.000 --> 00:00:08.000
From our test server.`))
	}))
	defer server.Close()

	service := &Service{}
	ctx := context.Background()

	result, err := service.fetchSubtitleText(ctx, server.URL)
	if err != nil {
		t.Fatalf("fetchSubtitleText() error = %v", err)
	}

	expected := "This is a test subtitle. From our test server."
	if result != expected {
		t.Errorf("fetchSubtitleText() = %q, want %q", result, expected)
	}
}

func TestFetchSubtitleTextEmptyURL(t *testing.T) {
	service := &Service{}
	ctx := context.Background()

	_, err := service.fetchSubtitleText(ctx, "")
	if err == nil {
		t.Error("fetchSubtitleText() with empty URL should return error")
	}
}

func TestFetchSubtitleTextTimeout(t *testing.T) {
	// Create a test server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than our context timeout
	}))
	defer server.Close()

	service := &Service{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := service.fetchSubtitleText(ctx, server.URL)
	if err == nil {
		t.Error("fetchSubtitleText() with timeout should return error")
	}
}

func TestFetchSubtitleText404(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	service := &Service{}
	ctx := context.Background()

	_, err := service.fetchSubtitleText(ctx, server.URL)
	if err == nil {
		t.Error("fetchSubtitleText() with 404 should return error")
	}
}

// Note: Integration test for generateSummary would require more complex mocking
// The main functionality is tested through the individual subtitle parsing and fetching tests above
