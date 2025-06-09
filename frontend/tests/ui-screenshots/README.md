# UI Screenshot Tests

This directory contains automated UI screenshot tests for the YouTube Curator v2 frontend. These tests support both **visual regression testing** and **screenshot generation** for PR reviews.

## Overview

The screenshot tests capture images of all major pages and UI states:

- **Home Page**: Videos list in light/dark mode, empty state
- **Subscriptions Page**: Channel management in light/dark mode, empty state
- **Notifications Page**: SMTP configuration and newsletter management  
- **Responsive Views**: Mobile (iPhone SE) and tablet (iPad) layouts

## Two Modes of Operation

### 1. Visual Regression Testing (Default)
Compares current UI against committed baseline screenshots and **fails tests** if changes are detected.

```bash
# Run visual regression tests
npm run test:screenshots

# Update baselines when UI changes are intentional
npm run test:screenshots:update
```

### 2. Screenshot Generation Only
Generates fresh screenshots for PR review **without** comparison - tests never fail due to UI changes.

```bash
# Generate screenshots without comparison
npm run test:screenshots:generate

# Or use environment variable
DISABLE_VISUAL_REGRESSION=true npm run test:screenshots
```

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

### Initial Setup

Generate baseline screenshots for visual regression testing:
```bash
npm run test:screenshots:update
```

This creates baseline images in `tests/ui-screenshots/ui-screenshots.spec.ts-snapshots/`.

### Local Development

```bash
# Visual regression testing (default)
npm run test:screenshots

# Screenshot generation only (for PR review)
npm run test:screenshots:generate

# Run tests with browser visible (headed mode)
npm run test:screenshots:headed

# Run tests with UI mode for debugging
npm run test:screenshots:ui

# Update baseline screenshots after UI changes
npm run test:screenshots:update
```

### CI Behavior

- Screenshot tests only run on **Pull Requests**, not on pushes to main
- Currently configured for **visual regression mode** (fails on UI changes)
- Generated screenshots are uploaded as GitHub Actions artifacts
- To switch to screenshot generation mode, set `DISABLE_VISUAL_REGRESSION: true` in the workflow

## Switching Between Modes

You can easily switch between visual regression testing and screenshot generation:

### For Visual Regression Testing
1. Ensure baseline screenshots exist: `npm run test:screenshots:update`
2. Set `DISABLE_VISUAL_REGRESSION: false` in `.github/workflows/run-tests.yml`
3. Tests will fail if UI changes vs baselines

### For Screenshot Generation Only  
1. Set `DISABLE_VISUAL_REGRESSION: true` in `.github/workflows/run-tests.yml`
2. Tests generate screenshots but never fail due to UI changes
3. Perfect for rapid UI development and PR review

## Test Structure

- `mock-data.ts`: Mock API responses for consistent test data
- `ui-screenshots.spec.ts`: Main test file with configurable screenshot capture
- `ui-screenshots.spec.ts-snapshots/`: Baseline screenshots directory (auto-generated)

## Configuration

The tests use:
- Mock API responses for reliable, consistent UI states
- Dark mode toggling for theme testing
- Responsive viewport testing for mobile/tablet views
- Network idle waiting to ensure page stability
- Configurable behavior via `DISABLE_VISUAL_REGRESSION` environment variable
- **Production-like mode**: Development tools (React Query DevTools, Next.js dev tools) are automatically disabled during testing for clean screenshots

### Environment Variables

- `DISABLE_VISUAL_REGRESSION=true`: Switch to screenshot generation mode (no comparison)
- `NEXT_PUBLIC_DISABLE_DEVTOOLS=true`: Disable React Query DevTools and other development tools (automatically set during testing)

### Development Tools Disabled During Testing

To ensure clean screenshots without development interference, the following are automatically disabled:
- **React Query DevTools**: The floating dev panel is hidden
- **Next.js Development Mode**: When running in CI, the app runs in production mode for clean output
- **Dev Server Indicators**: Any development-only UI elements are suppressed

### Font Consistency Across Platforms

To ensure consistent screenshot rendering between local development and CI environments:
- **Web Font**: The app uses Inter from Google Fonts as the primary font
- **Consistent Rendering**: Font loading is properly awaited before taking screenshots
- **Cross-Platform**: Same font renders identically on macOS, Windows, and Linux (CI)
- **Fallback Fonts**: System fonts are used as fallbacks if web fonts fail to load

This prevents font rendering differences between local screenshots and CI-generated screenshots.

## Artifacts

Screenshots are stored as GitHub Actions artifacts named `ui-screenshots` and retained for 14 days.

## Troubleshooting

### Font Rendering Consistency Between Local and CI

If you notice that locally generated screenshots don't match CI-generated screenshots due to font differences:

1. **Root Cause**: Different operating systems use different default fonts in the font stack
   - macOS: Uses -apple-system (San Francisco)
   - Windows: Uses Segoe UI
   - Linux (CI): Uses ui-sans-serif or Roboto

2. **Solution Applied**: The app now uses a prioritized font stack designed for cross-platform consistency:
   ```css
   font-family: ui-sans-serif, -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, 'Noto Sans', sans-serif;
   ```

3. **Browser Flags**: Playwright uses specific Chrome flags to ensure consistent font rendering:
   - `--font-render-hinting=none`: Disables font hinting variations
   - `--disable-font-subpixel-positioning`: Consistent character spacing
   - `--force-color-profile=srgb`: Consistent color rendering
   - `--disable-font-variations`: Prevents font weight variations

4. **Font Loading**: Tests wait for fonts to load completely before taking screenshots

If screenshots still differ between environments:
- Ensure you're using the same Node.js version locally as in CI (20.x)
- Clear your browser cache and Playwright cache: `npx playwright install --force`
- Regenerate baseline screenshots: `npm run test:screenshots:update`

### Other Issues

If tests fail:
1. **Visual Regression Mode**: Check if UI changes are intentional, then run `npm run test:screenshots:update`
2. Check if the development server is running correctly
3. Verify mock data matches current API response types
4. Ensure UI elements haven't changed (selectors may need updates)
5. Check for timing issues - add more wait time if needed