# Youtube Curator v2 Technical Plan

## 1. Introduction

This document outlines the technical architecture and implementation details for the Youtube Curator v2 project, a self-hosted platform for receiving email notifications of new YouTube videos from subscribed channels.

## 2. Architecture Overview

The system will follow a client-server architecture, with a Go backend handling data fetching, processing, scheduling, and email sending, and a Next.js frontend providing a web-based user interface for configuration.

*   **Backend (Go/Echo):** Responsible for core logic, scheduling, data storage interaction, and communication with external services (YouTube RSS, SMTP server).
*   **Frontend (Next.js/Tailwind):** A server-rendered React application providing the user interface for managing channels and settings.
*   **Database (BadgerDB):** An embedded key-value store used for persistent storage of channel subscriptions, user configurations, and state.
*   **Scheduler:** Integrated within the Go backend to trigger periodic checks of YouTube channels.

## 3. Data Flow

1.  User interacts with the Next.js frontend to add/remove channels or modify settings.
2.  Frontend communicates with the Go backend API to persist changes in BadgerDB.
3.  The scheduler in the Go backend triggers periodic checks.
4.  Backend fetches RSS feeds for subscribed channels.
5.  Backend parses RSS feeds, identifies new videos based on stored state (e.g., last checked video ID/timestamp).
6.  Backend compiles an email digest of new videos.
7.  Backend sends the email using configured SMTP details.

## 4. Component Breakdown

*   **Backend (Go):**
    *   API Endpoints (Echo framework) for frontend communication (add channel, remove channel, update settings, get status, etc.)
    *   RSS Fetcher: Module to fetch and parse YouTube RSS feeds.
    *   Scheduler: Module to manage the timing and execution of channel checks.
    *   Database Interface: Logic to interact with BadgerDB.
    *   Email Sender: Module to format and send emails via SMTP.
    *   Configuration Loader: Handles loading application configuration.
*   **Frontend (Next.js):**
    *   Pages for managing channels, configuring settings (SMTP, interval).
    *   Components for displaying channel lists, input forms.
    *   API interaction logic to communicate with the Go backend.
*   **Database (BadgerDB):**
    *   Store channel list (Channel ID, added timestamp).
    *   Store user settings (SMTP details, recipient email, check interval).
    *   Store per-channel state (Last checked video ID/timestamp).

## 5. Technology Stack Details

*   **Backend:** Go, Echo web framework.
*   **Frontend:** Next.js, React, Tailwind CSS.
*   **Database:** BadgerDB (embedded key-value store).
*   **Scheduling:** Go `time` and `sync` packages, potentially custom implementation or a simple library if needed.
*   **Deployment:** Docker.

## 6. YouTube RSS Integration Details

*   Utilize the format `https://www.youtube.com/feeds/videos.xml?channel_id=CHANNEL_ID`.
*   Need to handle extracting `CHANNEL_ID` from various YouTube URL formats.
*   Parsing requires handling the `media:` namespace for thumbnail, description, and content URLs.
*   Be aware of the limit of 10-15 recent videos per feed.
*   Implement strategies to handle potential rate limits (e.g., staggering requests, retries with backoff).
*   Initial sync strategy: When a channel is added, decide whether to fetch *all* available recent videos (up to the 10-15 limit) or only start tracking videos published *after* the addition.

## 7. Configuration Management

*   Sensitive information (SMTP passwords, etc.) should be handled securely.
*   Consider using environment variables or a dedicated configuration file (excluded from source control) for sensitive data in the Docker setup.
*   The UI will allow users to input and update some configuration values which will be stored in BadgerDB.

## 8. Error Handling and Monitoring

*   Implement robust error handling for external calls (RSS fetching, SMTP).
*   Logging will be crucial for diagnosing issues in a self-hosted environment.
*   Consider structured logging.
*   Basic monitoring could involve exposing a simple status endpoint or health check within the Go application.

## 9. Open Questions & Future Work

*   Refine initial data sync strategy for newly added channels.
*   Detailed error logging and reporting mechanism.
*   Implement secure storage for SMTP credentials.
*   Explore seamless migration to YouTube Data API if RSS proves insufficient or unreliable. 