package summary

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockService_BasicFunctionality(t *testing.T) {
	service := NewMockService(nil) // nil store is fine for mock

	ctx := context.Background()
	result := service.GetOrGenerateSummary(ctx, "test123")

	assert.NotNil(t, result)
	assert.NoError(t, result.Error)
	assert.Equal(t, "test123", result.VideoID)
	assert.NotEmpty(t, result.Summary)
	assert.Equal(t, "en", result.SourceLanguage)
	assert.False(t, result.Tracked)
}
