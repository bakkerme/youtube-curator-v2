# UI Screenshot Tests

This directory contains automated UI screenshot tests for the YouTube Curator v2 frontend. These tests run only on Pull Requests to provide visual feedback on UI changes.

## Overview

The screenshot tests capture images of all major pages and UI states:

- **Home Page**: Videos list in light/dark mode, empty state, search/filter states
- **Subscriptions Page**: Channel management in light/dark mode, empty state, import modal
- **Notifications Page**: SMTP configuration and newsletter results
- **Responsive Views**: Mobile and tablet layouts

## Running Tests

### Prerequisites

1. Install dependencies:
   ```bash
   npm install
   ```

2. Install Playwright browsers:
   ```bash
   npx playwright install chromium --with-deps
   ```

### Local Development

```bash
# Run all screenshot tests
npm run test:screenshots

# Run tests with browser visible (headed mode)
npm run test:screenshots:headed

# Run tests with UI mode for debugging
npm run test:screenshots:ui
```

### CI Behavior

- Screenshot tests only run on **Pull Requests**, not on pushes to main
- Generated screenshots are uploaded as GitHub Actions artifacts
- Tests include both light and dark mode variations
- Full page screenshots are captured with proper wait times

## Test Structure

- `mock-data.ts`: Mock API responses for consistent test data
- `ui-screenshots.spec.ts`: Main test file with all screenshot scenarios

## Configuration

The tests use:
- Mock API responses for reliable, consistent UI states
- Dark mode toggling for theme testing
- Responsive viewport testing for mobile/tablet views
- Network idle waiting to ensure page stability

## Artifacts

Screenshots are stored as GitHub Actions artifacts named `ui-screenshots` and retained for 14 days.

## Troubleshooting

If tests fail:
1. Check if the development server is running correctly
2. Verify mock data matches current API response types
3. Ensure UI elements haven't changed (selectors may need updates)
4. Check for timing issues - add more wait time if needed