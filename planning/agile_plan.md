# YouTube Curator v2 Agile Plan

## 1. Introduction

This document outlines an agile-style plan for the development of YouTube Curator v2, focusing on delivering functionality in iterations, starting with a Minimum Viable Product (MVP) and progressing through enhanced features including AI-powered video analysis and interactive chat capabilities.

## 2. Minimum Viable Product (MVP) - Iteration 1 (DONE)

The goal of the MVP is to create a functional core system that can fetch videos from a predefined set of channels and send email notifications, demonstrating the end-to-end data flow.

**Definition of Done for MVP:**

*   The application can successfully fetch RSS feeds from a list of YouTube channels provided via environment variables.
*   The application can identify new videos published since the last check.
*   The application can compile a basic email containing information about the new videos.
*   The application can send this email using SMTP credentials and a recipient address provided via environment variables.
*   The application persists enough state (e.g., the timestamp or ID of the last checked video per channel) to correctly identify new videos on subsequent runs.
*   The core logic is runnable, configured via environment variables.

**Key Tasks for MVP Iteration:**

*   Set up basic Go project structure.
*   Implement logic to read channel list, SMTP details (server, port, username, password), and recipient email from environment variables.
*   Implement RSS fetching logic for channels read from environment variables.
*   Implement RSS parsing to extract video information (title, link, published date). (Use https://github.com/bakkerme/ai-news-processor/tree/main/internal/rss as the basis)
*   Implement logic to compare current feed with last known state and identify new videos.
*   Integrate BadgerDB for storing the last checked state per channel.
*   Implement basic email formatting for new video notifications.
*   Implement email sending functionality using SMTP settings read from environment variables.
*   Refactor fetching, parsing, and email sending into a scheduled job (using Go's timer).
*   Ensure the application runs correctly when configured via environment variables.


## 3. Iteration 2: Dockerization and Deployment (DONE)

The goal of this iteration is to package the MVP into a Docker container for easy self-hosting and initial testing.

**Definition of Done for Iteration 2:** 

*   A Dockerfile is created to build a container for the Go backend (MVP logic).
*   The Docker container is configured to run the MVP application.
*   Basic configuration (hardcoded for MVP) is handled within the Docker environment.
*   Instructions are provided for building and running the MVP Docker container.
*   BadgerDB data persistence in a container is addressed (e.g., volume mounting).

**Key Tasks for Iteration 2:**

*   Create a Dockerfile for the Go backend application.
*   Configure the Dockerfile to include necessary dependencies and compile the Go code.
*   Define how BadgerDB data will persist outside the container (e.g., using a Docker volume).
*   Write basic documentation for building and running the Docker container.
*   Verify the MVP runs correctly within the Docker container.

## 4. Iteration 3: User Configuration & Basic Channel UI (DONE)

The goal of this iteration is to enable users to configure channels and the check interval via a basic web UI, persisting these settings.

**Definition of Done for Iteration 3:**

*   Backend API endpoints exist for adding, listing, and removing channels.
*   Backend API endpoints exist for setting and retrieving the check interval.
*   BadgerDB schema updated to store user-configured channels and interval.
*   Basic Next.js project structure set up.
*   Frontend pages/components created for viewing subscribed channels, adding a new channel (via ID or URL), and removing a channel.
*   Frontend page/component created for setting the check interval.
*   Frontend communicates with backend APIs to manage channels and the interval.
*   The scheduler reads the check interval from the database.
*   Docker Compose file updated to include the Next.js frontend service.
*   Docker configuration updated to allow communication between frontend and backend containers.

## 5. Iteration 4: User Configuration & Basic SMTP UI (DONE)

This iteration focuses on adding the UI and backend logic for users to configure their SMTP details.

**Definition of Done for Iteration 4:**

*   Backend API endpoints exist for setting and retrieving SMTP configuration.
*   BadgerDB schema updated to store SMTP settings.
*   Frontend page/component created for inputting and saving SMTP details (server, port, username, password, recipient email).
*   Frontend communicates with backend APIs to manage SMTP settings.
*   The email sender uses the configured SMTP details from the database.

**Key Tasks for Iteration 4:**

*   Define and implement backend API routes using Echo for SMTP configuration.
*   Implement storage mechanism for SMTP credentials in BadgerDB
*   Develop the Next.js frontend page/component for SMTP configuration.
*   Implement frontend logic to interact with the backend SMTP API.
*   Modify the email sender module to read SMTP details from BadgerDB.

## 6. Iteration 5: Latest Videos UI and Search (DONE)

The goal of this iteration is to provide a user interface to view all fetched videos from subscribed channels, with search and pagination capabilities.

**Definition of Done for Iteration 5:**

*   Backend API endpoint exists to retrieve a paginated and searchable list of all videos from subscribed channels.
*   BadgerDB schema updated to store video metadata (title, link, published date, channel) for display.
*   Frontend page/component created to display a grid or list of videos, as per the provided UI concept.
*   Frontend implements search functionality to filter videos by title or channel, and a filter for "Today" videos.
*   Frontend implements pagination for navigating through the video list.
*   Frontend communicates with the new backend API to fetch and display video metadata.

**Key Tasks for Iteration 5:**

*   Define and implement backend API routes using Echo for video metadata retrieval, including support for filtering by "Today" and pagination.
*   Modify the RSS processing logic to persist comprehensive video metadata (including thumbnail URLs and descriptions) in BadgerDB.
*   Implement search and pagination logic via the frontend
*   Develop the Next.js frontend page/component for video display (e.g., `pages/videos.tsx`).
*   Create reusable React components for video cards, search input, and pagination controls, including the "Today" filter button.
*   Implement frontend logic to interact with the backend video API, handling search queries and page changes.

## 7. Iteration 6:  Robustness, Error Handling, and Initial Sync (DONE)

Improve the system's reliability, add necessary logging and monitoring, and define/implement the initial sync behavior for new channels.

**Definition of Done for Iteration 6:**

*   Comprehensive error handling implemented for RSS fetching, parsing, database operations, and email sending.
*   Structured logging implemented throughout the application.
*   Basic monitoring/health check endpoint available in the backend.
*   Initial sync strategy for newly added channels is defined and implemented (e.g., only notify for videos *after* addition).
*   Graceful handling of RSS feed limitations and potential rate limits implemented (e.g., retries, backoff).

**Key Tasks for Iteration 6:**

*   Add detailed error handling and propagation in backend modules.
*   Integrate a structured logging library (if not using standard library) and add logs at key points.
*   Implement a simple HTTP endpoint in the Go backend for health checking.
*   Refine the channel adding logic to mark a timestamp or video ID upon addition for initial sync.
*   Modify the video identification logic to respect the initial sync strategy.
*   Implement retry logic with exponential backoff for RSS feed fetching errors.

## 8. Iteration 7: yt-dlp Integration for Enhanced Metadata and Comments

Integrate yt-dlp to augment existing RSS feed data with richer video metadata, comments, and subtitles for future AI analysis.

**Definition of Done for Iteration 7:**

*   yt-dlp command-line tool integrated into Go backend for video metadata, comments, and subtitles extraction.
*   Enhanced video metadata collection (duration, view count, tags, detailed descriptions) using yt-dlp.
*   Comment and subtitle retrieval functionality implemented for individual videos using yt-dlp.
*   Backend API endpoints updated to serve enhanced video metadata, comments, and subtitles to frontend.
*   Frontend updated to display additional video information (duration, view count) and provide access to comments/subtitles.

**Key Tasks for Iteration 7:**

*   Implement logic to execute yt-dlp commands from the Go backend.
*   Parse yt-dlp JSON output to extract comprehensive video metadata.
*   Extend BadgerDB schema to store enhanced video metadata, comments, and subtitles.
*   Update video processing pipeline to fetch data using yt-dlp in addition to RSS feeds.
*   Implement comment and subtitle extraction and storage from yt-dlp output.
*   Update frontend components to display enhanced metadata and provide access to comments and subtitles.

## 9. Iteration 8: LLM Integration and Comment Analysis

Integrate Large Language Model services and implement AI-driven comment and video analysis functionality, including summarization for email notifications.

**Definition of Done for Iteration 8:**

*   LLM service integration implemented with provider abstraction (OpenAI primary).
*   Comment and video analysis pipeline created for processing yt-dlp extracted data.
*   AI-generated insights include sentiment analysis, theme extraction, and summarization for both video content and comments.
*   Analysis results stored in BadgerDB with efficient caching.
*   Backend API endpoints created for retrieving analysis data.
*   Email notification system updated to include AI-generated video and comment summaries.
*   Configuration interface for LLM API credentials and settings.

**Key Tasks for Iteration 8:**

*   Implement LLM client with OpenAI API integration and provider abstraction.
*   Create comment and video analysis pipeline with batch processing capabilities.
*   Develop AI prompt templates for video content, comment sentiment, and theme analysis.
*   Implement analysis result caching and storage in BadgerDB.
*   Create backend API endpoints for analysis data retrieval.
*   Modify email generation logic to incorporate AI-generated summaries.
*   Add LLM configuration page to settings interface.
*   Implement error handling and graceful degradation for AI service failures.

## 10. Iteration 9: Basic Video Playback UI and Summary Visualization

Implement the foundational video playback interface with embedded YouTube videos, and integrate the visualization of AI-generated summary data.

**Definition of Done for Iteration 9:**

*   Video playback page created with clean, distraction-free YouTube video embedding.
*   YouTube Player API integrated for enhanced video control.
*   Navigation from video list to playback page implemented.
*   Basic video information displayed alongside the player (title, channel, description).
*   AI-generated video and comment summaries are displayed on the playback page.
*   Responsive design ensuring proper playback and summary display on mobile and desktop.
*   Video URL routing and deep linking functionality implemented.

**Key Tasks for Iteration 9:**

*   Create dedicated video playback page in Next.js.
*   Integrate YouTube Player API for embedded video functionality.
*   Implement video routing with URL parameters for direct video access.
*   Design responsive video player layout with Tailwind CSS.
*   Add basic video metadata display alongside the player.
*   Develop UI components to display AI-generated video and comment summaries on the playback page.
*   Implement navigation controls and breadcrumb functionality.

## 11. Iteration 10: Interactive Video Chat System

Implement AI-powered chat functionality for interactive discussions about video content, leveraging AI analysis.

**Definition of Done for Iteration 10:**

*   Interactive chat interface integrated into video playback page.
*   Context-aware AI chat system that understands current video content and its analysis.
*   Chat conversation history persistence in BadgerDB.
*   Real-time or near-real-time chat responses with loading states.
*   Chat system can answer questions about video content, creator, analysis insights, and related topics.
*   Conversation history management with session persistence.
*   Mobile-responsive chat interface with proper keyboard handling.

**Key Tasks for Iteration 10:**

*   Design and implement chat UI components with conversation history.
*   Create context-aware chat engine that includes video metadata and AI analysis insights in AI prompts.
*   Implement conversation history storage and retrieval system.
*   Add real-time communication (WebSocket or Server-Sent Events) for chat responses.
*   Create chat context management system for video-specific conversations.
*   Implement mobile-responsive chat interface with proper UX patterns.
*   Add chat conversation export and management features.

## 12. Iteration 11: AI Feature Enhancement and Optimization

Enhance AI capabilities with advanced analysis features and performance optimizations.

**Definition of Done for Iteration 11:**

*   Advanced comment analysis including trend detection and controversial topic identification.
*   Batch processing optimization for multiple videos and large comment datasets.
*   AI usage monitoring and cost tracking for self-hosted environments.
*   Enhanced chat capabilities with multi-turn conversation context.
*   Performance optimizations for AI API calls and response caching.
*   User preference system for AI analysis depth and chat behavior.

**Key Tasks for Iteration 11:**

*   Implement advanced comment analysis algorithms (trend detection, controversy scoring).
*   Optimize batch processing for improved performance and reduced API costs.
*   Create AI usage monitoring and cost tracking dashboard.
*   Enhance chat system with improved context management and conversation flow.
*   Implement intelligent caching strategies for AI responses.
*   Add user preference controls for AI feature customization.
*   Create comprehensive error handling and retry mechanisms for AI services.

## 13. Iteration 12: Polish, Testing, and Production Readiness

Focus on user experience improvements, comprehensive testing, and production deployment preparation.

**Definition of Done for Iteration 12:**

*   Comprehensive testing suite covering all AI features and video playback functionality.
*   Performance testing and optimization for self-hosted environments.
*   Enhanced error handling and user feedback for all AI interactions.
*   Documentation updated to include AI setup and configuration instructions.
*   Security audit completed for AI credential storage and user data handling.
*   Production deployment guide with AI service configuration examples.

**Key Tasks for Iteration 12:**

*   Implement comprehensive unit and integration tests for AI features.
*   Conduct performance testing and optimization for video playback and chat systems.
*   Create detailed documentation for AI setup and configuration.
*   Perform security audit focusing on API credential management.
*   Optimize Docker configuration for production deployment with AI services.
*   Create troubleshooting guides for common AI integration issues.

## 14. Future Iterations & Backlog

Features and improvements beyond the initial AI-enhanced product increments:

### Advanced AI Features
*   **Video Transcript Analysis:** AI-powered video content summarization and key point extraction.
*   **Cross-Video Intelligence:** Pattern recognition across multiple videos from the same creator.
*   **Automated Tagging:** AI-generated tags and categories for better video organization.
*   **Content Recommendations:** AI-driven suggestions based on viewing history and preferences.

### Enhanced User Experience
*   **Advanced Playback Features:** Playlist creation, watch progress tracking, and bookmarking.
*   **Social Features:** Sharing analysis insights and chat conversations.
*   **Mobile App:** Dedicated mobile application with offline AI chat capabilities.
*   **Multi-language Support:** AI analysis and chat in multiple languages.

### Technical Enhancements
*   **Local LLM Support:** Integration with self-hosted language models for privacy.
*   **Advanced Analytics:** Machine learning for user behavior analysis and optimization.
*   **Plugin Architecture:** Extensible system for custom AI analysis modules.
*   **Real-time Collaboration:** Multi-user chat rooms and shared viewing experiences.

### Integration Improvements
*   **Multiple LLM Providers:** Support for Anthropic, Google, and other AI services.
*   **Advanced Scheduling:** AI-powered optimal notification timing.
*   **Enhanced Import/Export:** Bulk operations for channels and analysis data.
*   **API Access:** Public API for integrating with other applications and services.