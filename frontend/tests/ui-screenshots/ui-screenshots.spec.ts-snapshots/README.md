# UI Screenshots Baseline Directory

This directory contains baseline screenshots used for visual regression testing.

## How It Works

1. **Initial Setup**: Run `npm run test:screenshots:update` to generate baseline screenshots
2. **Regular Testing**: Run `npm run test:screenshots` to compare current UI against baselines
3. **Updating Baselines**: When UI changes are intentional, run `npm run test:screenshots:update` to update baselines

## Directory Structure

- `home-videos-light.png` - Home page with videos (light mode)
- `home-videos-dark.png` - Home page with videos (dark mode)
- `home-empty-light.png` - Home page empty state
- `subscriptions-channels-light.png` - Subscriptions page with channels (light mode)
- `subscriptions-channels-dark.png` - Subscriptions page with channels (dark mode)
- `subscriptions-empty-light.png` - Subscriptions page empty state
- `notifications-light.png` - Notifications page (light mode)
- `notifications-dark.png` - Notifications page (dark mode)
- `mobile-home-light.png` - Mobile view of home page
- `tablet-subscriptions-light.png` - Tablet view of subscriptions page

## Switching Modes

### Visual Regression Testing (Default)
Tests fail if UI changes vs baselines:
```bash
npm run test:screenshots
```

### Screenshot Generation Only
Generate screenshots without comparison (for PR review):
```bash
npm run test:screenshots:generate
```

Or set environment variable:
```bash
DISABLE_VISUAL_REGRESSION=true npm run test:screenshots
```

## Updating Baselines

When you make intentional UI changes:
```bash
npm run test:screenshots:update
```

This will update all baseline screenshots with the current UI state.