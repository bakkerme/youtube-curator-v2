package openai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"youtube-curator-v2/internal/customerrors"
	"youtube-curator-v2/internal/http/retry"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// SchemaParameters contains the schema-related parameters for chat completion
type SchemaParameters struct {
	Schema      interface{}
	Name        string
	Description string
}

// OpenAIClient defines the interface for interacting with an OpenAI-compatible API
type OpenAIClient interface {
	// ChatCompletion performs a general-purpose chat completion request
	// systemPrompt: The system prompt to use
	// userPrompts: A list of user messages to send
	// imageURLs: Optional list of image URLs to include in the prompt
	// schemaParams: Optional schema parameters for response formatting (can be nil)
	// temperature: The temperature to use for the API call
	// maxTokens: Optional max tokens parameter to limit the response length (0 means no limit)
	// returns: Channel that will receive the response or error
	ChatCompletion(
		systemPrompt string,
		userPrompts []string,
		imageURLs []string,
		schemaParams *SchemaParameters,
		temperature float64,
		maxTokens int,
		results chan customerrors.ErrorString,
	)

	// SetRetryConfig updates the retry behavior configuration
	SetRetryConfig(config retry.RetryConfig)

	// PreprocessYAML extracts YAML content from the API response
	PreprocessYAML(response string) string

	// PreprocessJSON extracts JSON content from the API response
	PreprocessJSON(response string) string

	// GetModelName returns the model name used by this client
	GetModelName() string
}

// DefaultOpenAIRetryConfig provides sensible default values for OpenAI retry behavior
var DefaultOpenAIRetryConfig = retry.RetryConfig{
	MaxRetries:      5,
	InitialBackoff:  1 * time.Second,
	MaxBackoff:      30 * time.Second,
	BackoffFactor:   2.0,
	MaxTotalTimeout: 30 * time.Minute, // LLM calls can take a while
}

type Client struct {
	client *openai.Client
	model  string
	retry  retry.RetryConfig
}

// New creates a new OpenAI client
func New(baseURL, key, model string) *Client {
	client := openai.NewClient(
		option.WithAPIKey(key),
		option.WithBaseURL(baseURL),
		option.WithJSONSet("cache_set", true),
	)
	return &Client{
		client: &client,
		model:  model,
		retry:  DefaultOpenAIRetryConfig,
	}
}

// isModelLoadingError checks if the error is specifically a 404 due to model loading
func isModelLoadingError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "404 Not Found") &&
		strings.Contains(errStr, "Failed to load model") &&
		strings.Contains(errStr, "Model does not exist")
}

// ChatCompletion sends a request to the OpenAI API with the given prompts, optional images, and schema
func (c *Client) ChatCompletion(
	ctx context.Context,
	systemPrompt string,
	userPrompts []string,
	imageURLs []string,
	schemaParams *SchemaParameters,
	temperature float64,
	maxTokens int,
	results chan customerrors.ErrorString,
) {
	// Prepare messages array
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	// If we have image URLs, create a message with multi-modal content
	if len(imageURLs) > 0 {
		// Build image content parts
		contentParts := []openai.ChatCompletionContentPartUnionParam{}

		// First, add a text part if we have userPrompts
		if len(userPrompts) > 0 {
			textPart := openai.TextContentPart(userPrompts[0]) // First prompt as the text part
			contentParts = append(contentParts, textPart)
		}

		// Then add all the image parts
		for _, imgURL := range imageURLs {
			if imgURL != "" { // Basic validation
				imageParam := openai.ChatCompletionContentPartImageImageURLParam{
					URL: imgURL,
					// Optional: Detail: openai.String("auto"), // Can be "low", "high", or "auto"
				}
				imagePart := openai.ImageContentPart(imageParam)
				contentParts = append(contentParts, imagePart)
			}
		}

		// Create a user message with the multi-modal content parts
		messages = append(messages, openai.UserMessage(contentParts))

		// If there are additional prompts (beyond the first one), add them separately
		if len(userPrompts) > 1 {
			// Join the remaining prompts and add as a separate message
			messages = append(messages, openai.UserMessage(strings.Join(userPrompts[1:], "\n")))
		}
	} else {
		// No images, just add text prompts as usual
		messages = append(messages, openai.UserMessage(strings.Join(userPrompts, "\n")))
	}

	currentTemperature := 1.0
	if temperature != 0.0 {
		currentTemperature = temperature
	}

	params := openai.ChatCompletionNewParams{
		Model:       c.model,
		Messages:    messages,
		Temperature: param.NewOpt(currentTemperature),
	}

	// Add max tokens parameter if it's greater than 0
	if maxTokens > 0 {
		params.MaxTokens = openai.Int(int64(maxTokens))
	}

	if schemaParams != nil {
		schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        schemaParams.Name,
			Description: openai.String(schemaParams.Description),
			Schema:      schemaParams.Schema,
			Strict:      openai.Bool(true),
		}
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: schemaParam},
		}
	}

	shouldRetry := func(err error) bool {
		return isModelLoadingError(err)
	}

	ChatCompletionFn := func(ctx context.Context) (*openai.ChatCompletion, error) {
		return c.client.Chat.Completions.New(ctx, params)
	}

	resp, err := retry.RetryWithBackoff(ctx, c.retry, ChatCompletionFn, shouldRetry)

	if err != nil {
		var errMsg string
		if isModelLoadingError(err) {
			errMsg = fmt.Sprintf("model failed to load after retries: %v", err)
		} else {
			errMsg = fmt.Sprintf("error during API call: %v", err)
		}

		results <- customerrors.ErrorString{
			Value: "",
			Err:   errors.New(errMsg),
		}
		return
	}

	if len(resp.Choices) == 0 {
		results <- customerrors.ErrorString{
			Value: "",
			Err:   fmt.Errorf("empty response from llm"),
		}
		return
	}

	results <- customerrors.ErrorString{
		Value: resp.Choices[0].Message.Content,
		Err:   nil,
	}
}

// PreprocessYAML extracts YAML content from the API response
func (c *Client) PreprocessYAML(response string) string {
	return preprocess(response, "yaml")
}

func (c *Client) PreprocessJSON(response string) string {
	return preprocess(response, "json")
}

// GetModelName returns the model name used by this client
func (c *Client) GetModelName() string {
	return c.model
}

// preprocess extracts content of the specified format from the API response
func preprocess(response, format string) string {
	// Remove think tags and their contents
	thinkStart := "<think>"
	thinkEnd := "</think>"
	for {
		startIdx := strings.Index(response, thinkStart)
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(response, thinkEnd)
		if endIdx == -1 {
			break
		}
		response = response[:startIdx] + response[endIdx+len(thinkEnd):]
	}

	// Find the start markers with various possible formats
	startMarkers := []string{"```" + format, "```\n" + format, "```\r\n" + format}
	endMarker := "```"

	// Try each possible start marker format
	for _, startMarker := range startMarkers {
		startIdx := strings.Index(response, startMarker)
		if startIdx != -1 {
			// Calculate content start position based on the marker
			contentStart := startIdx + len(startMarker)

			endIdx := strings.Index(response[contentStart:], endMarker)
			if endIdx == -1 {
				// If no end marker found, return from start marker to end
				return strings.TrimSpace(response[contentStart:])
			}

			// Extract the content between markers
			content := response[contentStart : contentStart+endIdx]
			return strings.TrimSpace(content)
		}
	}

	// Handle case where just the format name appears on a line (possibly with whitespace)
	lines := strings.Split(response, "\n")
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == format && i < len(lines)-1 {
			// Found a line with just the format name, content starts from next line
			contentStart := strings.Join(lines[i+1:], "\n")

			endIdx := strings.Index(contentStart, endMarker)
			if endIdx == -1 {
				// If no end marker found, return everything
				return strings.TrimSpace(contentStart)
			}

			// Extract content up to the end marker
			content := contentStart[:endIdx]
			return strings.TrimSpace(content)
		}
	}

	// If no start marker found, return the original string trimmed
	return strings.TrimSpace(response)
}

// SetRetryConfig updates the retry configuration
func (c *Client) SetRetryConfig(config retry.RetryConfig) {
	c.retry = config
}
