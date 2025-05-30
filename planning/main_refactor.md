# Refactoring Plan for `checkForNewVideos` in `main.go`

## Problem

The `checkForNewVideos` function currently has multiple responsibilities, including iterating through channels, fetching RSS feeds, interacting with the database to get/set timestamps, finding new videos, and sending emails. This makes it difficult to test in isolation, as tests would require mocking the database, RSS feed provider, and email sender simultaneously, and controlling the complex internal loop logic.

## Proposed Solution

Refactor the `checkForNewVideos` function by extracting the channel-specific processing logic into a separate, testable component. This will be achieved by:

1.  **Defining an Interface:** Create a new Go interface, e.g., `ChannelProcessor`, that defines a method responsible for processing a single channel (e.g., `ProcessChannel`). This method will handle fetching the feed for a given channel ID, comparing timestamps, identifying the latest new video, and updating the last checked timestamp in the database for that channel.
2.  **Creating a Concrete Implementation:** Create a type, e.g., `DefaultChannelProcessor`, that implements the `ChannelProcessor` interface. The actual logic for processing a single channel (currently within the loop in `checkForNewVideos`) will be moved into the method of this type.
3.  **Modifying `checkForNewVideos`:** Update the `checkForNewVideos` function to accept a `ChannelProcessor` interface as a dependency. The function will then iterate through the configured channels and delegate the processing of each individual channel to the provided `ChannelProcessor` instance. It will collect the results (e.g., latest new videos) from each channel's processing and handle the final steps, such as aggregating the new videos and sending a combined email.
4.  **Updating `main`:** In the `main` function, an instance of the concrete `DefaultChannelProcessor` will be created and passed to `checkForNewVideos`.

## Benefits

*   **Improved Testability:** By introducing the `ChannelProcessor` interface, the logic for processing a single channel can be tested independently. Furthermore, the `checkForNewVideos` function can be tested by providing a mock implementation of `ChannelProcessor`, allowing focused testing of the aggregation and email sending logic without external dependencies like the actual database or RSS feed fetches.
*   **Separation of Concerns:** The refactoring separates the concerns of processing individual channels from the concern of orchestrating the overall check, aggregating results, and sending the final email.
*   **Increased Flexibility:** Using an interface allows for potential future alternative implementations of channel processing logic if needed (e.g., a different strategy for handling errors or fetching feeds).

This plan aims to make the core video checking logic more modular and easier to test and maintain.

## Implementation Status

✅ **Completed** - The refactoring has been successfully implemented:

1. **Interface Definition**: Created `ChannelProcessor` interface in `internal/processor/channel_processor.go`
2. **Concrete Implementation**: Implemented `DefaultChannelProcessor` that encapsulates the channel processing logic
3. **Refactored main.go**: Updated `checkForNewVideos` to accept and use `ChannelProcessor` interface
4. **Updated main function**: Modified to create and inject `DefaultChannelProcessor` instance

### Key Improvements Achieved:

- **Enhanced Testability**: Created comprehensive unit tests for `ChannelProcessor` that mock the database and RSS feed provider
- **Created main tests**: Added tests for `checkForNewVideos` using mock implementations, demonstrating isolated testing
- **Clear Separation of Concerns**: Channel processing logic is now separate from orchestration logic
- **Dependency Injection**: Using interfaces allows for easy mocking and testing

### Files Added/Modified:
- Added: `internal/processor/channel_processor.go` - Contains the interface and implementation
- Added: `internal/processor/channel_processor_test.go` - Unit tests for the processor
- Added: `main_test.go` - Tests for the refactored `checkForNewVideos` function
- Modified: `main.go` - Refactored to use the new `ChannelProcessor` interface