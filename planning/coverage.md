Here is the current test coverage:

```
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
        youtube-curator-v2              coverage: 0.0% of statements
ok      youtube-curator-v2/internal/config      0.003s  coverage: 81.6% of statements
ok      youtube-curator-v2/internal/email       0.020s  coverage: 35.1% of statements
        youtube-curator-v2/internal/http/retry          coverage: 0.0% of statements
ok      youtube-curator-v2/internal/rss 0.019s  coverage: 35.2% of statements
        youtube-curator-v2/internal/store               coverage: 0.0% of statements
youtube-curator-v2/internal/config/config.go:27:        LoadConfig              76.3%
youtube-curator-v2/internal/config/config.go:103:       loadChannelsFromFile    100.0%
youtube-curator-v2/internal/email/email.go:27:          NewEmailSender          0.0%
youtube-curator-v2/internal/email/email.go:38:          Send                    0.0%
youtube-curator-v2/internal/email/email.go:79:          FormatNewVideosEmail    68.4%
youtube-curator-v2/internal/http/retry/retry.go:96:     RetryWithBackoff        0.0%
youtube-curator-v2/internal/http/retry/retry.go:180:    IsRateLimitError        0.0%
youtube-curator-v2/internal/http/retry/retry.go:185:    GetRetryAfterDuration   0.0%
youtube-curator-v2/internal/rss/entry.go:25:            FeedString              100.0%
youtube-curator-v2/internal/rss/entry.go:74:            String                  100.0%
youtube-curator-v2/internal/rss/entry.go:88:            GetID                   100.0%
youtube-curator-v2/internal/rss/entry.go:93:            GetContent              100.0%
youtube-curator-v2/internal/rss/entry.go:98:            UnmarshalXML            91.7%
youtube-curator-v2/internal/rss/entry.go:126:           UnmarshalXML            100.0%
youtube-curator-v2/internal/rss/feedprovider.go:20:     NewFeedProvider         0.0%
youtube-curator-v2/internal/rss/feedprovider.go:25:     FetchFeed               0.0%
youtube-curator-v2/internal/rss/internal_funcs.go:29:   fetchRSS                0.0%
youtube-curator-v2/internal/rss/internal_funcs.go:44:   processRSSFeed          100.0%
youtube-curator-v2/internal/rss/internal_funcs.go:54:   CleanContent            100.0%
youtube-curator-v2/internal/rss/internal_funcs.go:82:   fetchWithRetry          0.0%
youtube-curator-v2/internal/rss/internal_funcs.go:134:  dumpFeed                0.0%
youtube-curator-v2/internal/rss/mock_provider.go:16:    NewMockFeedProvider     0.0%
youtube-curator-v2/internal/rss/mock_provider.go:23:    FetchFeed               0.0%
youtube-curator-v2/internal/store/store.go:16:          NewStore                0.0%
youtube-curator-v2/internal/store/store.go:26:          Close                   0.0%
youtube-curator-v2/internal/store/store.go:31:          GetLastCheckedVideoID   0.0%
youtube-curator-v2/internal/store/store.go:50:          SetLastCheckedVideoID   0.0%
youtube-curator-v2/internal/store/store.go:58:          GetLastCheckedTimestamp 0.0%
youtube-curator-v2/internal/store/store.go:78:          SetLastCheckedTimestamp 0.0%
youtube-curator-v2/main.go:20:                          main                    0.0%
youtube-curator-v2/main.go:81:                          checkForNewVideos       0.0%
total:                                                  (statements)            26.5%
```

### Module: youtube-curator-v2 (main package)
Coverage: 0.0% (main, checkForNewVideos)

**Problem:** The `main` function is typically difficult to test directly as it handles application bootstrapping. The `checkForNewVideos` function likely orchestrates interactions between several modules (store, RSS feed provider, email sender), making it hard to test in isolation without external dependencies.

**Proposal:**
1. **Refactor `checkForNewVideos`:** Extract the core logic of checking for videos, interacting with the store, fetching feeds, and sending emails into smaller, independent functions within a dedicated struct or package (e.g., a `Curator` service).
2. **Dependency Injection:** Design the `Curator` struct (or equivalent) to accept interfaces for its dependencies (Store, FeedProvider, EmailSender) via its constructor. This allows substituting real implementations with mock objects during testing.
3. **Test Extracted Logic:** Write unit tests for the smaller functions and the `Curator` struct using mock dependencies to verify interactions and logic without hitting the actual database, network, or email service.

**Information Needed for Modification:**
- The current implementation of `main` and `checkForNewVideos`.
- Definition of any types used or returned by `checkForNewVideos`.
- The interfaces or types used by `checkForNewVideos` for interacting with the store, RSS feeds, and email sending.

### Module: internal/http/retry
Coverage: 0.0% (RetryWithBackoff, IsRateLimitError, GetRetryAfterDuration)

**Problem:** The `RetryWithBackoff` function likely involves making actual HTTP calls, which are non-deterministic and have external dependencies, making direct unit testing challenging. The helper functions (`IsRateLimitError`, `GetRetryAfterDuration`) depend on the structure of `http.Response` which might be difficult to construct realistically in tests without making real calls.

**Proposal:**
1. **Abstract HTTP Calls:** Introduce an interface for performing HTTP requests. The `retry` package would then depend on this interface rather than directly using the `net/http` package.
2. **Test `RetryWithBackoff` with Mocks:** Implement a mock HTTP client that satisfies the new interface. Write unit tests for `RetryWithBackoff` that use this mock client to simulate various scenarios (success on first try, multiple retries, rate limits, different backoff behaviors) without making actual network calls.
3. **Test Helper Functions:** Write unit tests for `IsRateLimitError` and `GetRetryAfterDuration` by manually creating `http.Response` objects with different status codes and headers to simulate various server responses and verify the functions' logic.

**Information Needed for Modification:**
- The current implementation of `RetryWithBackoff`, `IsRateLimitError`, and `GetRetryAfterDuration`.
- How the `RetryWithBackoff` function is used and what type of HTTP client it currently uses.
- The expected structure of `http.Response` for rate limit errors.

### Module: internal/store
Coverage: 0.0% (NewStore, Close, GetLastCheckedVideoID, SetLastCheckedVideoID, GetLastCheckedTimestamp, SetLastCheckedTimestamp)

**Problem:** The `store` module appears to interact with a database (`youtubecurator.db`), making it difficult to test without setting up and managing a database instance. Direct testing of database interactions can be slow and require complex setup/teardown.

**Proposal:**
1. **Introduce a Store Interface:** Define an interface (`Store`) that outlines the methods for interacting with the storage (e.g., `GetLastCheckedVideoID`, `SetLastCheckedVideoID`, etc.).
2. **Implement the Interface:** The current database-backed implementation would satisfy this `Store` interface.
3. **Create a Mock Store:** Implement a separate struct that also satisfies the `Store` interface but uses in-memory data structures (like maps) instead of a database.
4. **Use Mock Store for Unit Tests:** Write unit tests for components that *use* the `Store` interface by injecting the mock store implementation. This allows testing the logic of those components quickly and reliably without database dependencies.
5. **Integration Tests for Database Store:** Write a separate set of integration tests specifically for the database-backed `store` implementation. These tests would require a real database (possibly an in-memory SQLite database for testing ease) to ensure the database interactions work correctly.

**Information Needed for Modification:**
- The current implementation of the `store` package, including how it connects to and interacts with the database.
- The methods currently available in the `store` package.
- The specific database library or driver being used.

### Module: internal/email
Coverage: 35.1% (NewEmailSender, Send have 0.0% coverage)

**Problem:** The `NewEmailSender` likely involves setup related to an external email service. The `Send` function directly interacts with an external email sending mechanism, which is an external dependency and should not be hit during unit tests.

**Proposal:**
1. **Define an Email Sender Interface:** Create an interface (`EmailSender`) with a `Send` method (and possibly other methods like `NewEmailSender` if appropriate, although `NewEmailSender` is often better as a constructor outside the interface).
2. **Implement the Interface:** The current email sending logic would implement this `EmailSender` interface.
3. **Create a Mock Email Sender:** Implement a mock struct that satisfies the `EmailSender` interface but instead of sending emails, it stores the email details in memory (e.g., a slice of sent email structs).
4. **Inject Mock Sender:** Components that need to send emails should be refactored to accept an `EmailSender` interface via dependency injection.
5. **Test with Mock:** Write unit tests for the components that send emails, injecting the mock sender. Verify that the mock sender's `Send` method was called with the expected email content and recipients.

**Information Needed for Modification:**
- The current implementation of the `email` package, including how `NewEmailSender` is used and how the `Send` function works.
- The external email sending library or service being used.
- The structure of the email data being sent (recipients, subject, body, etc.).

### Module: internal/rss
Coverage: 35.2% (NewFeedProvider, FetchFeed, fetchRSS, fetchWithRetry, dumpFeed, NewMockFeedProvider, FetchFeed in mock_provider.go have 0.0% coverage)

**Problem:** The `rss` module interacts with external RSS feed URLs (`FetchFeed`, `fetchRSS`, `fetchWithRetry`), which are external network dependencies. The mock provider itself is not being tested, indicating the system under test likely isn't using the mock in a test scenario.

**Proposal:**
1. **Define a Feed Provider Interface:** Create an interface (`FeedProvider`) with a method like `FetchFeed(url string) (*Feed, error)`.
2. **Implement Real Provider:** Create a struct (e.g., `HTTPFeedProvider`) that implements the `FeedProvider` interface and uses the `net/http` package (potentially with the `retry` logic) to fetch RSS feeds over the network.
3. **Test HTTPFeedProvider with Mock HTTP Client:** Use a mock HTTP client (as suggested for the `retry` module) to test the `HTTPFeedProvider` without making real network calls. Simulate different HTTP responses (success, errors, invalid XML).
4. **Implement Mock Provider:** Create a separate struct (e.g., `MockFeedProvider`) that also implements the `FeedProvider` interface but returns predefined mock `Feed` data based on the input URL.
5. **Inject FeedProvider Interface:** Refactor components that use the RSS feed provider to accept a `FeedProvider` interface via dependency injection.
6. **Test Components with Mock Provider:** Write unit tests for components that consume RSS feeds, injecting the `MockFeedProvider`. This allows testing how the system handles different feed contents and errors without external network requests.
7. **Test Mock Provider:** Write tests specifically for the `MockFeedProvider` to ensure it correctly returns the predefined mock data when its `FetchFeed` method is called with expected URLs.

**Information Needed for Modification:**
- The current implementation of the `rss` package, including how feeds are fetched and processed.
- The structure of the `Feed` and `Entry` types used.
- How the `MockFeedProvider` is intended to be used.