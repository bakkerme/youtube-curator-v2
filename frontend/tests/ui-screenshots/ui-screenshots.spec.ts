import { test, expect, Page } from '@playwright/test';
import { mockVideoEntries, mockChannels, mockSMTPConfig, emptyMockData } from './mock-data';
import * as path from 'path';

// Configuration: set DISABLE_VISUAL_REGRESSION=true to generate screenshots without comparison
const DISABLE_VISUAL_REGRESSION = process.env.DISABLE_VISUAL_REGRESSION === 'true';

// Test only runs on PRs, not on regular builds
test.describe('UI Screenshots for PR Review', () => {
  // Setup API mocking before each test
  test.beforeEach(async ({ page }) => {
    // Mock Next.js config API route
    await page.route('**/api/config', async (route) => {
      await route.fulfill({ 
        json: { 
          apiUrl: 'http://localhost:8080/api' 
        } 
      });
    });

    // Mock backend API responses - use the backend API URLs
    await page.route('**/localhost:8080/api/videos*', async (route) => {
      await route.fulfill({ 
        json: { 
          videos: mockVideoEntries, 
          lastRefresh: new Date().toISOString(),
          totalCount: mockVideoEntries.length 
        } 
      });
    });

    await page.route('**/localhost:8080/api/channels*', async (route) => {
      await route.fulfill({ json: { channels: mockChannels } });
    });

    await page.route('**/localhost:8080/api/config/smtp*', async (route) => {
      await route.fulfill({ json: mockSMTPConfig });
    });

    await page.route('**/localhost:8080/api/newsletter/run*', async (route) => {
      await route.fulfill({ 
        json: { 
          message: 'Newsletter run completed successfully',
          channelsProcessed: 5, 
          channelsWithError: 0,
          newVideosFound: 3,
          emailSent: true 
        }
      });
    });

    // Mock video watch endpoint
    await page.route('**/localhost:8080/api/videos/**/watch', async (route) => {
      await route.fulfill({ json: { success: true } });
    });
  });

  // Helper function to handle screenshots - supports both visual regression and generation modes
  async function captureScreenshot(page: Page, filename: string, options: any = {}) {
    if (DISABLE_VISUAL_REGRESSION) {
      // Screenshot generation mode - save to artifacts without comparison
      const screenshotsDir = path.join('test-results', 'screenshots');
      const fs = require('fs');
      if (!fs.existsSync(screenshotsDir)) {
        fs.mkdirSync(screenshotsDir, { recursive: true });
      }
      await page.screenshot({ 
        path: path.join(screenshotsDir, filename),
        fullPage: options.fullPage || true 
      });
    } else {
      // Visual regression mode - compare against baselines
      await expect(page).toHaveScreenshot(filename, {
        fullPage: true,
        timeout: 15000,
        ...options
      });
    }
  }

  // Helper function to wait for page load and React hydration
  async function waitForPageLoad(page: Page) {
    await page.waitForLoadState('networkidle');
    
    // Wait for web fonts to load completely
    await page.evaluate(() => {
      return new Promise<void>((resolve) => {
        if (document.fonts && document.fonts.ready) {
          document.fonts.ready.then(() => resolve());
        } else {
          // Fallback timeout if document.fonts is not available
          setTimeout(() => resolve(), 1000);
        }
      });
    });
    
    // Wait for React to hydrate and any loading states to complete
    await page.waitForTimeout(2000); 
  }

  test.describe('Home Page (Videos) Screenshots - Light Mode', () => {
    test.use({ colorScheme: 'light' });

    test('should capture home page with videos - light mode', async ({ page }) => {
      await page.goto('/');
      await waitForPageLoad(page);
      
      // Wait for videos to load by checking for video cards
      await page.waitForSelector('[data-testid="video-card"], .video-card, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'home-videos-light.png');
    });

    test('should capture home page - empty state - light mode', async ({ page }) => {
      // Mock empty videos response
      await page.route('**/localhost:8080/api/videos*', async (route) => {
        await route.fulfill({ 
          json: { 
            videos: [], 
            lastRefresh: new Date().toISOString(),
            totalCount: 0 
          } 
        });
      });
      
      await page.goto('/');
      await waitForPageLoad(page);
      
      await captureScreenshot(page, 'home-empty-light.png');
    });
  });

  test.describe('Home Page (Videos) Screenshots - Dark Mode', () => {
    test.use({ colorScheme: 'dark' });

    test('should capture home page with videos - dark mode', async ({ page }) => {
      await page.goto('/');
      await waitForPageLoad(page);
      
      // Wait for videos to load
      await page.waitForSelector('[data-testid="video-card"], .video-card, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'home-videos-dark.png');
    });

    test('should capture home page - empty state - dark mode', async ({ page }) => {
      // Mock empty videos response
      await page.route('**/localhost:8080/api/videos*', async (route) => {
        await route.fulfill({ 
          json: { 
            videos: [], 
            lastRefresh: new Date().toISOString(),
            totalCount: 0 
          } 
        });
      });
      
      await page.goto('/');
      await waitForPageLoad(page);
      
      await captureScreenshot(page, 'home-empty-dark.png');
    });
  });

  test.describe('Subscriptions Page Screenshots - Light Mode', () => {
    test.use({ colorScheme: 'light' });

    test('should capture subscriptions page with channels - light mode', async ({ page }) => {
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      // Wait for the page title to ensure content is loaded
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'subscriptions-channels-light.png');
    });

    test('should capture subscriptions page - empty state - light mode', async ({ page }) => {
      // Mock empty channels response
      await page.route('**/localhost:8080/api/channels', async (route) => {
        await route.fulfill({ json: { channels: [] } });
      });
      
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'subscriptions-empty-light.png');
    });
  });

  test.describe('Subscriptions Page Screenshots - Dark Mode', () => {
    test.use({ colorScheme: 'dark' });

    test('should capture subscriptions page with channels - dark mode', async ({ page }) => {
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'subscriptions-channels-dark.png');
    });

    test('should capture subscriptions page - empty state - dark mode', async ({ page }) => {
      // Mock empty channels response
      await page.route('**/localhost:8080/api/channels', async (route) => {
        await route.fulfill({ json: { channels: [] } });
      });
      
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'subscriptions-empty-dark.png');
    });
  });

  test.describe('Notifications/Settings Page Screenshots - Light Mode', () => {
    test.use({ colorScheme: 'light' });

    test('should capture notifications page - light mode', async ({ page }) => {
      await page.goto('/notifications');
      await waitForPageLoad(page);
      
      // Wait for forms to load
      await page.waitForSelector('form, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'notifications-light.png');
    });
  });

  test.describe('Notifications/Settings Page Screenshots - Dark Mode', () => {
    test.use({ colorScheme: 'dark' });

    test('should capture notifications page - dark mode', async ({ page }) => {
      await page.goto('/notifications');
      await waitForPageLoad(page);
      
      await page.waitForSelector('form, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'notifications-dark.png');
    });
  });

  test.describe('Responsive Design Screenshots - Light Mode', () => {
    test.use({ colorScheme: 'light' });

    test('should capture mobile view - home page', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/');
      await waitForPageLoad(page);
      
      await page.waitForSelector('[data-testid="video-card"], .video-card, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'mobile-home-light.png');
    });

    test('should capture tablet view - subscriptions page', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 }); // iPad
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'tablet-subscriptions-light.png');
    });
  });

  test.describe('Responsive Design Screenshots - Dark Mode', () => {
    test.use({ colorScheme: 'dark' });

    test('should capture mobile view - home page - dark mode', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/');
      await waitForPageLoad(page);
      
      await page.waitForSelector('[data-testid="video-card"], .video-card, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'mobile-home-dark.png');
    });

    test('should capture tablet view - subscriptions page - dark mode', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 }); // iPad
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'tablet-subscriptions-dark.png');
    });
  });

  test.describe('Watch Page Screenshots - Light Mode', () => {
    test.use({ colorScheme: 'light' });

    test('should capture watch page with valid video - light mode', async ({ page }) => {
      await page.goto('/watch/dQw4w9WgXcQ');
      await waitForPageLoad(page);
      
      // Wait for video title to load
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'watch-video-light.png');
    });

    test('should capture watch page with video not found - light mode', async ({ page }) => {
      await page.goto('/watch/nonexistent-video-id');
      await waitForPageLoad(page);
      
      // Wait for "Video Not Found" message
      await page.waitForSelector('h1:has-text("Video Not Found")', { timeout: 10000 });
      
      await captureScreenshot(page, 'watch-not-found-light.png');
    });

    test('should capture watch page loading state - light mode', async ({ page }) => {
      // Delay the API response to capture loading state
      await page.route('**/localhost:8080/api/videos*', async (route) => {
        await new Promise(resolve => setTimeout(resolve, 2000));
        await route.fulfill({ 
          json: { 
            videos: mockVideoEntries, 
            lastRefresh: new Date().toISOString(),
            totalCount: mockVideoEntries.length 
          } 
        });
      });
      
      await page.goto('/watch/dQw4w9WgXcQ');
      
      // Wait for loading skeleton to appear
      await page.waitForSelector('.animate-pulse', { timeout: 5000 });
      
      await captureScreenshot(page, 'watch-loading-light.png');
    });
  });

  test.describe('Watch Page Screenshots - Dark Mode', () => {
    test.use({ colorScheme: 'dark' });

    test('should capture watch page with valid video - dark mode', async ({ page }) => {
      await page.goto('/watch/dQw4w9WgXcQ');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'watch-video-dark.png');
    });

    test('should capture watch page with video not found - dark mode', async ({ page }) => {
      await page.goto('/watch/nonexistent-video-id');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1:has-text("Video Not Found")', { timeout: 10000 });
      
      await captureScreenshot(page, 'watch-not-found-dark.png');
    });

    test('should capture watch page loading state - dark mode', async ({ page }) => {
      // Delay the API response to capture loading state
      await page.route('**/localhost:8080/api/videos*', async (route) => {
        await new Promise(resolve => setTimeout(resolve, 2000));
        await route.fulfill({ 
          json: { 
            videos: mockVideoEntries, 
            lastRefresh: new Date().toISOString(),
            totalCount: mockVideoEntries.length 
          } 
        });
      });
      
      await page.goto('/watch/dQw4w9WgXcQ');
      
      await page.waitForSelector('.animate-pulse', { timeout: 5000 });
      
      await captureScreenshot(page, 'watch-loading-dark.png');
    });
  });

  test.describe('Watch Page Responsive Screenshots - Light Mode', () => {
    test.use({ colorScheme: 'light' });

    test('should capture mobile view - watch page - light mode', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/watch/abc123def45');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'mobile-watch-light.png');
    });

    test('should capture tablet view - watch page - light mode', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 }); // iPad
      await page.goto('/watch/xyz789ghi01');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'tablet-watch-light.png');
    });
  });

  test.describe('Watch Page Responsive Screenshots - Dark Mode', () => {
    test.use({ colorScheme: 'dark' });

    test('should capture mobile view - watch page - dark mode', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/watch/abc123def45');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'mobile-watch-dark.png');
    });

    test('should capture tablet view - watch page - dark mode', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 }); // iPad
      await page.goto('/watch/xyz789ghi01');
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'tablet-watch-dark.png');
    });
  });
});
