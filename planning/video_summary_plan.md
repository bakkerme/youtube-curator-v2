# Video Summarization Feature: Technical Plan

## 1. Overview

This document outlines the technical plan for implementing a video summarization feature. The feature will download video subtitles using `yt-dlp`, send the subtitles to an LLM via an OpenAPI-compatible endpoint, and expose the generated summary through a new API endpoint.

## 2. API Endpoint Specifications

### 2.1. Public Video Summary Endpoint

*   **Endpoint:** `GET /videos/{videoID}/summary`
*   **Description:** Retrieves a previously generated summary for the specified video. If a summary is not available, it triggers the generation process.
*   **Method:** `GET`
*   **Path Parameters:**
    *   `videoID` (string, required): The unique identifier of the YouTube video.
*   **Success Responses:**
    *   `200 OK`: Summary successfully retrieved.
        *   Content-Type: `application/json`
        *   Body:
            ```json
            {
              "videoId": "string",
              "summary": "string",
              "sourceLanguage": "string", // e.g., "en", "es"
              "generatedAt": "datetime", // ISO 8601 format
              "tracked": "bool"
            }
            ```
    *   `200 OK`: Summary successfully retrieved for arbitrary video.
        *   Content-Type: `application/json`
        *   Body:
            ```json
            {
              "videoId": "string",
              "summary": "string",
              "sourceLanguage": "string", // e.g., "en", "es"
              "generatedAt": "datetime", // ISO 8601 format
              "tracked": false
            }
            ```
*   **Error Responses:**
    *   `404 Not Found`: The video ID is not found, or subtitles are unavailable for the video.
    *   `500 Internal Server Error`: An unexpected error occurred during summary generation or retrieval (e.g., LLM service unavailable, `yt-dlp` error).
    *   `503 Service Unavailable`: The summary generation is in progress, and the client should retry later (optional, could also be a blocking call that waits for generation).

### 2.2. LLM Configuration Endpoint

*   **Endpoint:** `/config/llm`
*   **Description:** Manages the configuration for connecting to an OpenAPI-compatible LLM service. This includes setting or updating the API endpoint, API key, model, and other relevant parameters.
*   **Methods:**
    *   `GET`: Retrieves the current LLM configuration (excluding sensitive data like API keys if necessary, or only accessible by admins).
    *   `POST` or `PUT`: Sets or updates the LLM configuration.
*   **Request Body (for `POST`/`PUT`):**
    *   Content-Type: `application/json`
    *   Body:
        ```json
        {
          "endpointUrl": "string", // e.g., "https://api.llmprovider.com/v1/chat/completions"
          "apiKey": "string", // Handled securely, potentially stored in a secret manager
          "model": "string", // e.g., "gpt-3.5-turbo"
        }
        ```
*   **Success Responses:**
    *   `GET 200 OK`: Current configuration retrieved.
        *   Content-Type: `application/json`
        *   Body (example, API key might be masked or omitted):
            ```json
            {
              "endpointUrl": "string",
              "model": "string",
            }
            ```
    *   `POST/PUT 200 OK` or `204 No Content`: Configuration updated successfully.
*   **Error Responses:**
    *   `400 Bad Request`: Invalid request body or parameters.
    *   `401 Unauthorized` / `403 Forbidden`: If the endpoint requires authentication/authorization.
    *   `500 Internal Server Error`: Failed to save or process the configuration.

## 3. System Design and Flow

1.  **Request:** Client requests `GET /videos/{videoID}/summary`.
2.  **Cache Check:** The system checks the local cache/database for an existing summary for `videoID`.
    *   If found and valid, return the cached summary.
3.  **Subtitle Download:** If no valid summary exists:
    *   The system uses the existing `yt-dlp` integration to download subtitles for `videoID`.
        *   Priority will be given to user-selected language or auto-generated English subtitles if specific language is not available.
        *   If subtitles cannot be fetched, return `404 Not Found` or an appropriate error.
4.  **LLM Invocation:**
    *   The downloaded subtitle text is sent to the configured LLM service endpoint.
    *   The prompt will be engineered to request a concise, informative summary.
5.  **Summary Processing & Caching:**
    *   The LLM response (summary) is received.
    *   The summary, along with `videoID`, `sourceLanguage`, and `generatedAt` timestamp, is stored in the caching system.
6.  **Response:** The newly generated summary is returned to the client.

## 4. Updates to Video Caching System & Data Model

The existing `Entry` object, used throughout the system to store video data (as defined in `backend/internal/rss/entry.go`), will be extended to include the video summary and related metadata. This approach leverages the current data storage and caching mechanisms for `Entry` objects.

*   **`Entry` Struct Enhancements:**
    *   The `Entry` struct will be augmented with the following fields:
        * `Summary` struct
            *   `Summary` (string, omitempty): Stores the generated summary text.
            *   `SourceLanguage` (string, omitempty): The language of the subtitles used for the summary (e.g., "en", "es").
            *   `SummaryGeneratedAt` (time.Time, omitempty): Timestamp of when the summary was generated.
        *   These new fields will be included in the JSON representation of the `Entry` object.

*   **Data Persistence:**
    *   The existing mechanisms for serializing and persisting `Entry` objects (e.g., to `youtubecurator.db` or other storage) will automatically handle the new summary fields. No separate database table for summaries is required.
    *   The data access layer will need minor updates to ensure these new fields are correctly populated and retrieved.


## 5. Offline Mock System

To facilitate development and testing without actual external dependencies, mock systems for the `yt-dlp` enricher service and the LLM service will be implemented.

### 5.1. `yt-dlp` Enricher Service Mocking

*   **Mechanism:**
    *   Define a `MockEnricher` struct that implements the `ytdlp.Enricher` interface (defined in `backend/internal/ytdlp/enricher.go`).
    *   The application will use dependency injection to provide either the `DefaultEnricher` or the `MockEnricher` based on a configuration setting (e.g., an environment variable `MOCK_YTDLP_ENRICHER=true`).
*   **Functionality:**
    *   When the `MockEnricher` is active, its `EnrichEntry` method will be called instead of the one that executes the actual `yt-dlp` command.
    *   This mock `EnrichEntry` method will directly populate the `rss.Entry` object with predefined data, specifically focusing on providing mock subtitle information (e.g., by setting the `AutoSubtitles` field or a similar field designated for raw subtitle text).
    *   The mock subtitle data can be hardcoded, read from a configuration, or loaded from a local file structure (e.g., `/backend/feed_mocks/subtitles/{videoID}.txt`) based on the video ID from the `rss.Entry`.
    *   This approach allows developers to test the video summarization pipeline components that rely on subtitle data without needing to run `yt-dlp` or have actual video subtitles available.

### 5.2. LLM Service Mocking

*   **Mechanism:**
    *   Introduce a configuration setting for the LLM service endpoint URL. During development, this can be pointed to a local mock server.
    *   Alternatively, an environment variable (e.g., `DEBUG_MOCK_LLM=true`) can enable an in-process mock handler.
*   **Functionality:**
    *   The mock LLM endpoint will accept the same request payload as the real service.
    *   It will return a predefined static summary, or a simple transformation of the input (e.g., "Mock summary for video with subtitles starting: [first 50 chars of input text]...").
    *   This allows testing the flow of data to and from the LLM service without incurring costs or relying on network connectivity.
    *   Different mock responses can be configured based on parts of the input text or specific video IDs to test various LLM output scenarios.

## 6. Arbitrary Video Summarization

To allow summarization of arbitrary YouTube videos, the system will be updated to handle videos that are not part of the tracked channels. These videos will not be saved to the videos database, and the response will include `tracked: false`.

### 6.1. API Endpoint Behavior

* **Endpoint:** `GET /videos/{videoID}/summary`
* **Behavior for Arbitrary Videos:**
  * If the `videoID` does not belong to a tracked channel, the system will still attempt to generate a summary.
  * The generated summary will not be cached or stored in the database.
  * The response will include `tracked: false` to indicate that the video is not part of the tracked channels.

### 6.2. System Design and Flow Updates

1. **Request:** Client requests `GET /videos/{videoID}/summary`.
2. **Cache Check:**
   * If the `videoID` belongs to a tracked channel, the system checks the local cache/database for an existing summary.
   * If the `videoID` does not belong to a tracked channel, skip the cache check.
3. **Subtitle Download:**
   * The system uses the existing `yt-dlp` integration to download subtitles for `videoID`.
   * Priority will be given to user-selected language or auto-generated English subtitles if specific language is not available.
   * If subtitles cannot be fetched, return `404 Not Found` or an appropriate error.
4. **LLM Invocation:**
   * The downloaded subtitle text is sent to the configured LLM service endpoint.
   * The prompt will be engineered to request a concise, informative summary.
5. **Response:**
   * For tracked videos, the summary is cached and returned to the client.
   * For arbitrary videos, the summary is returned directly without caching, and the response includes `tracked: false`.

### 6.3. Updates to API Endpoint Specifications

#### Public Video Summary Endpoint

* **Success Responses:**
  * `200 OK`: Summary successfully retrieved.
    * Content-Type: `application/json`
    * Body:
      ```json
      {
        "videoId": "string",
        "summary": "string",
        "sourceLanguage": "string", // e.g., "en", "es"
        "generatedAt": "datetime", // ISO 8601 format
        "tracked": false
      }
      ```
* **Error Responses:**
  * `404 Not Found`: The video ID is not found, or subtitles are unavailable for the video.
  * `500 Internal Server Error`: An unexpected error occurred during summary generation or retrieval (e.g., LLM service unavailable, `yt-dlp` error).
  * `503 Service Unavailable`: The summary generation is in progress, and the client should retry later.