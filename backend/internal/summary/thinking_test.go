package summary

import (
	"testing"
)

func TestParseThinkingBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantThinking string
		wantSummary  string
	}{
		{
			name: "Response with thinking block",
			input: `<think>
This is my thinking process.
I need to analyze the video content.
</think>

This is the actual summary of the video content.`,
			wantThinking: "This is my thinking process.\nI need to analyze the video content.",
			wantSummary:  "This is the actual summary of the video content.",
		},
		{
			name: "Response without thinking block",
			input: "This is just a regular summary without any thinking.",
			wantThinking: "",
			wantSummary:  "This is just a regular summary without any thinking.",
		},
		{
			name: "Response with multiple thinking blocks",
			input: `<think>First thought process</think>

Some content here.

<think>Second thought process</think>

Final summary content.`,
			wantThinking: "First thought process\n\nSecond thought process",
			wantSummary:  "Some content here.\n\nFinal summary content.",
		},
		{
			name: "Response with markdown in summary",
			input: `<think>
Analyzing the video structure and content.
</think>

### **Summary of Cold Fusion Episode**

This episode covers:
- AI failures
- Software decline
- Internal chaos`,
			wantThinking: "Analyzing the video structure and content.",
			wantSummary:  "### **Summary of Cold Fusion Episode**\n\nThis episode covers:\n- AI failures\n- Software decline\n- Internal chaos",
		},
		{
			name: "Empty thinking block",
			input: `<think></think>

Regular summary content here.`,
			wantThinking: "",
			wantSummary:  "Regular summary content here.",
		},
		{
			name: "Thinking block with whitespace",
			input: `<think>
   
   Some thinking with extra whitespace   
   
</think>

Summary content.`,
			wantThinking: "Some thinking with extra whitespace",
			wantSummary:  "Summary content.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotThinking, gotSummary := parseThinkingBlocks(tt.input)
			
			if gotThinking != tt.wantThinking {
				t.Errorf("parseThinkingBlocks() thinking = %q, want %q", gotThinking, tt.wantThinking)
			}
			
			if gotSummary != tt.wantSummary {
				t.Errorf("parseThinkingBlocks() summary = %q, want %q", gotSummary, tt.wantSummary)
			}
		})
	}
}