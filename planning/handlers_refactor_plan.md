# API Handlers Refactoring Plan

## Overview

The current `backend/internal/api/handlers.go` file has grown to over 700 lines and contains multiple concerns that should be separated for better maintainability, testability, and code organization.

## Current State Analysis

### File Structure Breakdown
- **Lines 1-20**: Package imports and dependencies
- **Lines 21-150**: API request/response type definitions (~130 lines)
- **Lines 151-250**: Transformation functions (~100 lines)
- **Lines 251-280**: Handler struct and constructor (~30 lines)
- **Lines 281-700**: Handler method implementations (~420 lines)

### Identified Issues

1. **Single Responsibility Principle Violation**
   - Channel management (CRUD operations, import)
   - Configuration management (SMTP, LLM, intervals)
   - Video operations (listing, marking watched, summaries)
   - Newsletter operations (manual triggers)

2. **High Coupling**
   - Single handler struct with 8 dependencies
   - Mixed HTTP handling with business logic
   - Transformation logic embedded in handlers

3. **Poor Testability**
   - Large handler methods difficult to unit test
   - Multiple concerns make mocking complex
   - Business logic tightly coupled to HTTP layer

4. **Maintenance Challenges**
   - Changes to one domain affect the entire file
   - Difficult to locate specific functionality
   - Risk of merge conflicts in team environment

## Proposed Solution

### Phase 1: Domain Separation

Create domain-specific handler files with clear separation of concerns:

```
backend/internal/api/
├── handlers/
│   ├── base.go          # Shared handler dependencies and utilities
│   ├── channels.go      # Channel CRUD operations
│   ├── config.go        # Configuration management (SMTP, LLM, intervals)
│   ├── videos.go        # Video operations and summaries
│   └── newsletter.go    # Newsletter manual triggers
├── types/
│   ├── requests.go      # All API request types
│   ├── responses.go     # All API response types
│   └── transformers.go  # Type conversion functions
├── handlers.go          # Legacy file (to be removed after migration)
└── router.go           # Updated to use new handler structure
```

### Phase 2: Service Layer Extraction

Extract business logic into dedicated service layer:

```
backend/internal/services/
├── channel_service.go   # Channel business logic
├── config_service.go    # Configuration business logic
├── video_service.go     # Video business logic
└── newsletter_service.go # Newsletter business logic
```

### Phase 3: Dependency Injection Improvement

Create a service container to manage dependencies:

```go
// backend/internal/api/handlers/base.go
type ServiceContainer struct {
    Store          store.Store
    FeedProvider   rss.FeedProvider
    EmailSender    email.Sender
    Config         *config.Config
    Processor      processor.ChannelProcessor
    VideoStore     *store.VideoStore
    YtdlpEnricher  ytdlp.Enricher
    SummaryService summary.SummaryServiceInterface
}

type BaseHandler struct {
    services *ServiceContainer
}
```

## Implementation Plan

### Step 1: Create Type Files (Low Risk)
**Estimated Time**: 2-3 hours

1. Create `backend/internal/api/types/` directory
2. Move all request types to `requests.go`
3. Move all response types to `responses.go`
4. Move transformation functions to `transformers.go`
5. Update imports in `handlers.go`

**Files to Create**:
- `backend/internal/api/types/requests.go`
- `backend/internal/api/types/responses.go`
- `backend/internal/api/types/transformers.go`

### Step 2: Create Base Handler Structure (Medium Risk)
**Estimated Time**: 1-2 hours

1. Create `backend/internal/api/handlers/` directory
2. Create `base.go` with shared dependencies
3. Create service container structure

**Files to Create**:
- `backend/internal/api/handlers/base.go`

### Step 3: Extract Configuration Handlers (Low Risk)
**Estimated Time**: 3-4 hours

Start with configuration handlers as they have the least dependencies:
- `GetCheckInterval`
- `SetCheckInterval`
- `GetSMTPConfig`
- `SetSMTPConfig`
- `GetLLMConfig`
- `SetLLMConfig`

**Files to Create**:
- `backend/internal/api/handlers/config.go`

### Step 4: Extract Channel Handlers (Medium Risk)
**Estimated Time**: 4-5 hours

Extract channel management handlers:
- `GetChannels`
- `AddChannel`
- `RemoveChannel`
- `ImportChannels`

**Files to Create**:
- `backend/internal/api/handlers/channels.go`

### Step 5: Extract Video Handlers (Medium Risk)
**Estimated Time**: 4-5 hours

Extract video-related handlers:
- `GetVideos`
- `MarkVideoAsWatched`
- `GetVideoSummary`

**Files to Create**:
- `backend/internal/api/handlers/videos.go`

### Step 6: Extract Newsletter Handlers (Low Risk)
**Estimated Time**: 2-3 hours

Extract newsletter handlers:
- `RunNewsletter`

**Files to Create**:
- `backend/internal/api/handlers/newsletter.go`

### Step 7: Update Router (Medium Risk)
**Estimated Time**: 2-3 hours

Update `router.go` to use new handler structure:
- Create handler instances for each domain
- Update route definitions
- Ensure all endpoints are properly mapped

### Step 8: Testing and Cleanup (High Priority)
**Estimated Time**: 4-6 hours

1. Update existing tests to work with new structure
2. Add unit tests for each handler domain
3. Integration testing to ensure API compatibility
4. Remove original `handlers.go` file

## Risk Assessment

### Low Risk
- Type extraction (Step 1)
- Configuration handlers (Step 3)
- Newsletter handlers (Step 6)

### Medium Risk
- Base handler structure (Step 2)
- Channel handlers (Step 4)
- Video handlers (Step 5)
- Router updates (Step 7)

### High Risk
- Testing and cleanup (Step 8) - Critical for ensuring no regression

## Testing Strategy

### Unit Testing
- Test each handler domain independently
- Mock service dependencies
- Validate request/response transformations

### Integration Testing
- Test complete API endpoints
- Verify router configuration
- Ensure backward compatibility

### Migration Testing
- Run existing test suite after each step
- Compare API responses before/after migration
- Performance testing to ensure no degradation

## Success Criteria

1. **Maintainability**: Each handler file < 200 lines
2. **Testability**: >90% test coverage for handler logic
3. **Performance**: No regression in API response times
4. **Compatibility**: All existing API endpoints work unchanged
5. **Code Quality**: Pass all linting and static analysis

## Timeline

**Total Estimated Time**: 20-28 hours

- **Week 1**: Steps 1-3 (Type extraction and base structure)
- **Week 2**: Steps 4-6 (Handler extraction)
- **Week 3**: Steps 7-8 (Router updates and testing)

## Rollback Plan

If issues arise during migration:

1. **Immediate Rollback**: Keep original `handlers.go` until all tests pass
2. **Selective Rollback**: Each step can be independently rolled back
3. **Feature Flags**: Use build tags if needed to switch between implementations

## Benefits After Refactoring

1. **Single Responsibility**: Each handler file handles one domain
2. **Easier Testing**: Domain-specific handlers can be tested in isolation
3. **Better Maintainability**: Changes to video logic don't affect channel handlers
4. **Cleaner Router**: Clear separation of concerns in route definitions
5. **Reusable Business Logic**: Services can be used by other parts of the application
6. **Team Collaboration**: Reduced merge conflicts with smaller, focused files
7. **Future Extensibility**: Easy to add new endpoints within existing domains

## Code Examples

### Before (Current Structure)
```go
// Single large handlers.go file with everything
type Handlers struct {
    store          store.Store
    feedProvider   rss.FeedProvider
    emailSender    email.Sender
    config         *config.Config
    processor      processor.ChannelProcessor
    videoStore     *store.VideoStore
    ytdlpEnricher  ytdlp.Enricher
    summaryService summary.SummaryServiceInterface
}

func (h *Handlers) GetChannels(c echo.Context) error { ... }
func (h *Handlers) GetSMTPConfig(c echo.Context) error { ... }
func (h *Handlers) GetVideos(c echo.Context) error { ... }
func (h *Handlers) RunNewsletter(c echo.Context) error { ... }
```

### After (Proposed Structure)
```go
// backend/internal/api/handlers/base.go
type ServiceContainer struct {
    Store          store.Store
    FeedProvider   rss.FeedProvider
    EmailSender    email.Sender
    Config         *config.Config
    Processor      processor.ChannelProcessor
    VideoStore     *store.VideoStore
    YtdlpEnricher  ytdlp.Enricher
    SummaryService summary.SummaryServiceInterface
}

// backend/internal/api/handlers/channels.go
type ChannelHandlers struct {
    *BaseHandler
}

func (h *ChannelHandlers) GetChannels(c echo.Context) error { ... }
func (h *ChannelHandlers) AddChannel(c echo.Context) error { ... }

// backend/internal/api/handlers/config.go
type ConfigHandlers struct {
    *BaseHandler
}

func (h *ConfigHandlers) GetSMTPConfig(c echo.Context) error { ... }
func (h *ConfigHandlers) SetSMTPConfig(c echo.Context) error { ... }
```

## Detailed Handler Breakdown

### Current Handlers by Domain

#### Channel Handlers (4 methods)
- `GetChannels` - Retrieves all tracked channels
- `AddChannel` - Adds a new channel to track
- `RemoveChannel` - Removes a channel from tracking
- `ImportChannels` - Bulk import multiple channels

#### Configuration Handlers (6 methods)
- `GetCheckInterval` - Gets current check interval setting
- `SetCheckInterval` - Updates check interval setting
- `GetSMTPConfig` - Gets SMTP configuration (without password)
- `SetSMTPConfig` - Updates SMTP configuration
- `GetLLMConfig` - Gets LLM configuration (without API key)
- `SetLLMConfig` - Updates LLM configuration

#### Video Handlers (3 methods)
- `GetVideos` - Retrieves videos with optional refresh
- `MarkVideoAsWatched` - Marks a video as watched
- `GetVideoSummary` - Gets or generates video summary

#### Newsletter Handlers (1 method)
- `RunNewsletter` - Manually triggers newsletter generation

### Type Definitions to Extract

#### Request Types (8 types)
- `ChannelRequest`
- `ConfigInterval`
- `SMTPConfigRequest`
- `LLMConfigRequest`
- `ImportChannelsRequest`
- `ChannelImport`
- `RunNewsletterRequest`

#### Response Types (15 types)
- `ErrorResponse`
- `Pagination`
- `ChannelResponse`
- `ChannelsResponse`
- `VideoResponse` (and related nested types)
- `VideosResponse`
- `NewsletterRunResponse`
- `SMTPConfigResponse`
- `LLMConfigResponse`
- `VideoSummaryResponse`
- `ImportChannelsResponse`
- `ImportFailure`

#### Transformation Functions (8 functions)
- `transformChannel`
- `transformChannels`
- `transformVideoEntry`
- `transformVideoLink`
- `transformVideoAuthor`
- `transformVideoMediaGroup`
- `transformVideoMediaThumbnail`
- `transformVideoMediaContent`
- `transformVideos`

This refactoring will significantly improve the codebase's maintainability, testability, and development experience while maintaining full backward compatibility.
