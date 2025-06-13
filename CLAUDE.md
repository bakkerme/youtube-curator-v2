# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

YouTube Curator v2 is a self-hosted application that monitors YouTube channels via RSS feeds and sends email notifications when new videos are published. It consists of a Go backend with REST API and a Next.js frontend.

## Architecture

**Backend (Go)**
- **Entry Point**: `backend/main.go` - Contains main application logic, cron scheduling, and concurrent channel processing
- **API Layer**: `backend/internal/api/` - Echo-based REST API with handlers for channels, videos, config, and newsletter
- **Core Services**:
  - `internal/processor/` - Channel processing logic for RSS feeds
  - `internal/rss/` - RSS feed parsing and YouTube-specific parsing
  - `internal/store/` - BadgerDB storage layer for channels, videos, and configuration
  - `internal/email/` - SMTP email sending with HTML templates
  - `internal/summary/` - AI-powered video summarization using OpenAI
  - `internal/ytdlp/` - Video metadata enrichment via yt-dlp
- **Database**: BadgerDB (embedded key-value store) at `backend/youtubecurator.db/`

**Frontend (Next.js)**
- **Framework**: Next.js 15 with React 19, TypeScript, and Tailwind CSS
- **State Management**: TanStack Query for server state
- **Pages**: Home (videos), Subscriptions (channel management), Settings (notifications/config)
- **Components**: Video cards, pagination, configuration forms
- **Testing**: Jest + React Testing Library for unit tests, Playwright for E2E/visual regression

## Common Development Commands

**Backend (from project root):**
```bash
# Run backend in development
make run-backend

# Run backend with live reload
make run-backend-air

# Build backend binary
make build

# Run all Go tests
make test

# Clean database
make clean-db
```

**Frontend (from frontend/ directory):**
```bash
# Development server with Turbopack
npm run dev

# Run unit tests
npm test

# Run tests with coverage
npm test:coverage

# Run visual regression tests
npm run test:screenshots

# Build for production
npm run build

# Lint code
npm run lint
```

**Docker (from project root):**
```bash
# Start full stack with Docker Compose
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## Key Integration Points

- **API Communication**: Frontend communicates with backend via REST API at `/api/*` endpoints
- **Configuration Storage**: All settings (SMTP, LLM, intervals) stored in BadgerDB, accessible via API
- **Channel Management**: Channels stored in database, not config files (legacy channels.json still supported)
- **Video Processing**: Concurrent RSS processing with configurable worker pools
- **Email Templates**: HTML email templates in `backend/internal/email/templates/`

## Testing Strategy

- **Backend**: Standard Go testing with mocks for external dependencies (RSS feeds, email, OpenAI)
- **Frontend**: Jest for unit tests, MSW for API mocking, Playwright for E2E and screenshot testing
- **Integration**: API integration tests covering full request/response cycles

## Configuration Management

The application supports both environment variables and database-stored configuration:
- Environment variables for initial setup and Docker deployment
- Database storage for runtime configuration changes via the web UI
- SMTP, LLM, and scheduling settings are managed through the frontend settings page

## Important Notes

- The main scheduler logic is in `main.go` with concurrent channel processing
- RSS feeds are fetched concurrently with configurable worker pools (RSS_CONCURRENCY)
- Email notifications combine all new videos from a check cycle into a single email
- The application supports both interval-based and daily scheduling modes
- Mock providers exist for RSS feeds and other external services for testing

## Cursor Rules Integration

When making changes:
- Always run tests after writing them
- When proposing changes, describe the approach rather than immediately implementing
- Avoid changing implementation code when writing tests without explicit user request