package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"youtube-curator-v2/internal/api/types"
	"youtube-curator-v2/internal/store"
)

func TestGetNewsletterConfig(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     *store.NewsletterConfig
		mockError      error
		expectedStatus int
		expectedBody   types.NewsletterConfigResponse
	}{
		{
			name: "Success - Newsletter enabled",
			mockReturn: &store.NewsletterConfig{
				Enabled: true,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: types.NewsletterConfigResponse{
				Enabled: true,
			},
		},
		{
			name: "Success - Newsletter disabled",
			mockReturn: &store.NewsletterConfig{
				Enabled: false,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: types.NewsletterConfigResponse{
				Enabled: false,
			},
		},
		{
			name:           "Error - Store failure",
			mockReturn:     nil,
			mockError:      errors.New("store error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := store.NewMockStore(ctrl)
			baseHandlers := &BaseHandlers{store: mockStore}
			handler := NewConfigHandlers(baseHandlers)
			e := echo.New()

			// Setup mock
			mockStore.EXPECT().GetNewsletterConfig().Return(tt.mockReturn, tt.mockError)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/config/newsletter", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.GetNewsletterConfig(c)

			// Assert
			if tt.mockError != nil {
				assert.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, httpErr.Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				var response types.NewsletterConfigResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestSetNewsletterConfig(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    types.NewsletterConfigRequest
		mockError      error
		expectedStatus int
		expectedBody   types.NewsletterConfigResponse
	}{
		{
			name: "Success - Enable newsletter",
			requestBody: types.NewsletterConfigRequest{
				Enabled: true,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: types.NewsletterConfigResponse{
				Enabled: true,
			},
		},
		{
			name: "Success - Disable newsletter",
			requestBody: types.NewsletterConfigRequest{
				Enabled: false,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: types.NewsletterConfigResponse{
				Enabled: false,
			},
		},
		{
			name: "Error - Store failure",
			requestBody: types.NewsletterConfigRequest{
				Enabled: true,
			},
			mockError:      errors.New("store error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := store.NewMockStore(ctrl)
			baseHandlers := &BaseHandlers{store: mockStore}
			handler := NewConfigHandlers(baseHandlers)
			e := echo.New()

			// Setup mock
			expectedConfig := &store.NewsletterConfig{
				Enabled: tt.requestBody.Enabled,
			}
			mockStore.EXPECT().SetNewsletterConfig(expectedConfig).Return(tt.mockError)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/config/newsletter", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.SetNewsletterConfig(c)

			// Assert
			if tt.mockError != nil {
				assert.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, httpErr.Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				var response types.NewsletterConfigResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestSetNewsletterConfig_InvalidJSON(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockStore(ctrl)
	baseHandlers := &BaseHandlers{store: mockStore}
	handler := NewConfigHandlers(baseHandlers)
	e := echo.New()

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPut, "/api/config/newsletter", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.SetNewsletterConfig(c)

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}