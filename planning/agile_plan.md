# Youtube Curator v2 Agile Plan

## 1. Introduction

This document outlines an agile-style plan for the development of Youtube Curator v2, focusing on delivering functionality in iterations, starting with a Minimum Viable Product (MVP).

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

## 3. Iteration 2: Dockerization and Deployment (MVP) (DONE)

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

## 4. Iteration 3: User Configuration & Basic Channel UI

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

**Key Tasks for Iteration 3:**

*   Set up the basic Next.js project structure.
*   Define and implement backend API routes using Echo for channel management and interval settings.
*   Update BadgerDB access methods to handle dynamic channel lists and interval storage.
*   Develop the Next.js frontend pages/components for channel listing, adding, and removing.
*   Implement frontend logic to interact with the backend channel APIs.
*   Develop the Next.js frontend page/component for interval configuration.
*   Implement frontend logic to interact with the backend interval API.
*   Modify the scheduler to retrieve the interval from BadgerDB on startup and updates.
*   Implement logic in the backend to extract channel ID from various YouTube URLs provided by the user.
*   Update Dockerfile for the Go backend to expose the API port.
*   Create a Dockerfile for the Next.js frontend.
*   Develop/update a Docker Compose file to manage both backend and frontend containers.
*   Configure networking in Docker Compose for inter-container communication.

## 5. Iteration 4: User Configuration & Basic SMTP UI

This iteration focuses on adding the UI and backend logic for users to configure their SMTP details securely.

**Definition of Done for Iteration 4:**

*   Backend API endpoints exist for setting and retrieving SMTP configuration.
*   BadgerDB schema updated to store SMTP settings securely.
*   Frontend page/component created for inputting and saving SMTP details (server, port, username, password, recipient email).
*   Frontend communicates with backend APIs to manage SMTP settings.
*   The email sender uses the configured SMTP details from the database.
*   Docker configuration updated to handle sensitive SMTP credentials securely.

**Key Tasks for Iteration 4:**

*   Define and implement backend API routes using Echo for SMTP configuration.
*   Implement secure storage mechanism for sensitive SMTP credentials in BadgerDB (e.g., encryption or relying on filesystem permissions if applicable in Docker).
*   Develop the Next.js frontend page/component for SMTP configuration.
*   Implement frontend logic to interact with the backend SMTP API.
*   Modify the email sender module to read SMTP details from BadgerDB.
*   Update Docker configuration (e.g., using environment variables or Docker secrets) for passing SMTP credentials securely.

## 6. Iteration 5: Robustness, Error Handling, and Initial Sync

Improve the system's reliability, add necessary logging and monitoring, and define/implement the initial sync behavior for new channels.

**Definition of Done for Iteration 5:**

*   Comprehensive error handling implemented for RSS fetching, parsing, database operations, and email sending.
*   Structured logging implemented throughout the application.
*   Basic monitoring/health check endpoint available in the backend.
*   Initial sync strategy for newly added channels is defined and implemented (e.g., only notify for videos *after* addition).
*   Graceful handling of RSS feed limitations and potential rate limits implemented (e.g., retries, backoff).

**Key Tasks for Iteration 5:**

*   Add detailed error handling and propagation in backend modules.
*   Integrate a structured logging library (if not using standard library) and add logs at key points.
*   Implement a simple HTTP endpoint in the Go backend for health checking.
*   Refine the channel adding logic to mark a timestamp or video ID upon addition for initial sync.
*   Modify the video identification logic to respect the initial sync strategy.
*   Implement retry logic with exponential backoff for RSS feed fetching errors.

## 7. Future Iterations & Backlog

Features and improvements beyond the initial shippable product increments:

*   YouTube Data API Integration (as a potential alternative or supplement to RSS).
*   Advanced Scheduling Options.
*   Email Content Customization.
*   Import/Export Channel List.
*   More advanced UI Enhancements (searching, filtering, grouping).
*   Watch Page (embedding video into a web page). 