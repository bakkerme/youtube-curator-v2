# YouTube Curator v2 Technical Plan

## 1. Introduction

This document outlines the technical architecture and implementation details for the YouTube Curator v2 project, a self-hosted platform for receiving email notifications of new YouTube videos from subscribed channels, enhanced with AI-powered video analysis and interactive chat capabilities.

## 2. Architecture Overview

The system will follow a client-server architecture, with a Go backend handling data fetching, processing, scheduling, email sending, and AI integration, and a Next.js frontend providing a comprehensive web-based user interface for configuration, video viewing, and AI-powered interactions.

*   **Backend (Go/Echo):** Core logic, scheduling, data storage, external service communication (YouTube RSS/API, SMTP, LLM APIs).
*   **Frontend (Next.js/Tailwind):** Server-rendered React application with video playback, AI chat interface, and management tools.
*   **Database (BadgerDB):** Embedded key-value store for channels, configurations, video metadata, chat history, and analysis cache.
*   **Scheduler:** Integrated periodic YouTube channel checking and comment analysis processing.
*   **AI Services:** LLM integration for comment analysis and interactive chat functionality.

## 3. Enhanced Data Flow

### Core Data Flow
1.  User interacts with the Next.js frontend to add/remove channels or modify settings.
2.  Frontend communicates with the Go backend API to persist changes in BadgerDB.
3.  The scheduler triggers periodic checks and comment analysis.
4.  Backend fetches RSS feeds and enhanced metadata for subscribed channels.
5.  Backend processes new videos and triggers comment analysis.
6.  Backend compiles email digest and sends notifications.

### AI-Enhanced Flow
7.  User accesses video through the playback UI.
8.  Backend retrieves and analyzes YouTube comments using LLM services.
9.  Frontend displays embedded video with AI-generated comment insights.
10. User interacts with AI chat about video content.
11. Backend processes chat queries with video context and LLM integration.

## 4. Component Breakdown

### Backend (Go)
*   **API Endpoints (Echo):**
    *   Channel management (add, remove, list)
    *   Configuration management (SMTP, LLM settings, intervals)
    *   Video metadata and playback endpoints
    *   Comment analysis endpoints
    *   AI chat endpoints with conversation history
    *   Health check and status endpoints
*   **Core Modules:**
    *   RSS Fetcher: YouTube RSS feed processing
    *   YouTube API Client: Enhanced metadata and comment retrieval
    *   Scheduler: Channel checks and analysis processing
    *   Database Interface: BadgerDB operations with enhanced schema
    *   Email Sender: SMTP-based notifications
    *   Configuration Loader: Multi-source configuration management
*   **AI Integration:**
    *   LLM Client: OpenAI API or compatible service integration
    *   Comment Analyzer: Batch comment processing and insight generation
    *   Chat Engine: Context-aware conversation management
    *   Analysis Cache: Optimized storage and retrieval of AI-generated insights

### Frontend (Next.js)
*   **Pages:**
    *   Channel management dashboard
    *   Latest videos grid with search and filtering
    *   Enhanced video playback page with embedded player
    *   Configuration pages (SMTP, LLM, intervals)
    *   AI chat interface integrated with video viewing
*   **Components:**
    *   Video player wrapper with YouTube embedding
    *   Comment analysis display with sentiment and themes
    *   Interactive chat interface with conversation history
    *   Channel list management with bulk operations
    *   Search and filter controls with "Today" quick filter
    *   Responsive design for mobile and desktop viewing
*   **State Management:**
    *   Video playback state and progress tracking
    *   Chat conversation history and context
    *   Comment analysis caching and updates
    *   User preferences and UI settings

### Database Schema (BadgerDB)
*   **Existing Data:**
    *   Channel subscriptions (ID, metadata, last checked)
    *   User settings (SMTP, check intervals)
    *   Video tracking state (last processed IDs/timestamps)
    *   Video metadata (title, description, thumbnail, duration, view count)
*   **Enhanced Data:**
    *   Comment analysis cache (sentiment, themes, summary, timestamp)
    *   Chat conversation history (per-video and global contexts)
    *   LLM API configuration (provider, model, rate limits)
    *   User interaction analytics (viewing patterns, chat usage)

## 5. Technology Stack Details

*   **Backend:** Go, Echo web framework, YouTube Data API v3
*   **Frontend:** Next.js, React, Tailwind CSS, YouTube Player API
*   **Database:** BadgerDB with enhanced schema design
*   **AI Integration:** OpenAI API (primary), with abstraction for other LLM providers
*   **Scheduling:** Go `time` and `context` packages with enhanced job queue
*   **Deployment:** Docker with multi-stage builds and optimized images

## 6. YouTube Integration Details

### RSS Feed Integration
*   Format: `https://www.youtube.com/feeds/videos.xml?channel_id=CHANNEL_ID`
*   Handle various YouTube URL formats for channel ID extraction
*   Parse `media:` namespace for thumbnails, descriptions, and content URLs
*   Manage 10-15 video limit per feed with intelligent caching

### YouTube Data API Integration
*   **Video Details:** Enhanced metadata including duration, view counts, tags
*   **Comment Retrieval:** Top-level comments with threading support
*   **Rate Limit Management:** Intelligent quota usage and backoff strategies
*   **Batch Processing:** Efficient API calls for multiple videos
*   **Fallback Strategy:** Graceful degradation when API limits are reached

## 7. AI Integration Architecture

### LLM Service Integration
*   **Provider Abstraction:** Support for OpenAI, Anthropic, and local models
*   **API Configuration:** Secure credential management and endpoint configuration
*   **Rate Limiting:** Intelligent request batching and quota management
*   **Error Handling:** Graceful degradation and retry mechanisms

### Comment Analysis Pipeline
*   **Data Collection:** Batch retrieval of top comments per video
*   **Preprocessing:** Comment filtering, deduplication, and sanitization
*   **Analysis Processing:** Sentiment analysis, theme extraction, and summarization
*   **Caching Strategy:** Efficient storage and retrieval of analysis results
*   **Update Mechanism:** Periodic re-analysis of popular videos

### Interactive Chat System
*   **Context Management:** Video-specific and conversation history context
*   **Query Processing:** Natural language understanding and intent recognition
*   **Response Generation:** Context-aware responses with video information
*   **Conversation Persistence:** Chat history storage and retrieval
*   **Real-time Communication:** WebSocket or Server-Sent Events for live chat

## 8. Configuration Management

### Security Considerations
*   **API Key Management:** Secure storage for YouTube API and LLM credentials
*   **Environment Variables:** Docker secrets and encrypted configuration files
*   **User Data Protection:** Secure handling of viewing history and chat data
*   **Rate Limit Configuration:** Configurable quotas and usage monitoring

### Configurable Parameters
*   **LLM Settings:** Provider selection, model configuration, analysis depth
*   **Comment Analysis:** Processing frequency, comment count limits, analysis types
*   **Chat Behavior:** Response style, context window size, conversation limits
*   **UI Preferences:** Default views, analysis display options, chat interface settings

## 9. Performance Considerations

### Optimization Strategies
*   **Caching:** Multi-level caching for API responses, analysis results, and UI state
*   **Async Processing:** Background jobs for comment analysis and LLM interactions
*   **Resource Management:** Efficient memory usage for large comment datasets
*   **Database Optimization:** Indexed queries and batch operations for BadgerDB

### Scalability Planning
*   **Concurrent Processing:** Goroutine pools for parallel API calls and analysis
*   **Queue Management:** Background job processing for time-intensive operations
*   **Storage Efficiency:** Compressed storage for large analysis datasets
*   **API Usage Optimization:** Intelligent caching to minimize external API calls

## 10. Error Handling and Monitoring

### Enhanced Error Handling
*   **AI Service Failures:** Graceful degradation when LLM services are unavailable
*   **API Rate Limits:** Intelligent backoff and alternative processing strategies
*   **Data Corruption:** Recovery mechanisms for comment analysis and chat history
*   **Network Issues:** Retry logic with exponential backoff for all external services

### Monitoring and Observability
*   **Structured Logging:** Comprehensive logging for all AI interactions and processing
*   **Health Checks:** Enhanced endpoints for system and AI service status
*   **Usage Analytics:** Tracking of AI feature usage and performance metrics
*   **Cost Monitoring:** LLM API usage tracking and budget alerts

## 11. Open Questions & Future Work

### Technical Challenges
*   **LLM Cost Management:** How to optimize AI usage for self-hosted environments with budget constraints?
*   **Real-time Chat Performance:** What's the optimal balance between response quality and speed?
*   **Comment Analysis Scope:** How to determine the optimal number of comments to analyze per video?
*   **Data Privacy:** How to ensure user chat data remains private in self-hosted environments?
*   **Multi-language Support:** How to handle comment analysis for non-English YouTube content?

### Future Technical Enhancements
*   **Local LLM Support:** Integration with self-hosted language models
*   **Advanced Analytics:** Machine learning for viewing pattern analysis
*   **Real-time Notifications:** WebSocket-based live updates for new videos
*   **Mobile App:** React Native or Flutter app with offline capabilities
*   **Plugin Architecture:** Extensible system for custom analysis and chat modules