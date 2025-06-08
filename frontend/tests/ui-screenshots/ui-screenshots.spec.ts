import { test, expect, Page } from '@playwright/test';
import { mockVideoEntries, mockChannels, mockSMTPConfig, emptyMockData } from './mock-data';

// Test only runs on PRs, not on regular builds
test.describe('UI Screenshots for PR Review', () => {
  // Setup API mocking before each test
  test.beforeEach(async ({ page }) => {
    // Mock API responses
    await page.route('**/api/videos*', async (route) => {
      const url = route.request().url();
      if (url.includes('empty=true')) {
        await route.fulfill({ 
          json: { 
            videos: [], 
            lastRefresh: new Date().toISOString(),
            totalCount: 0 
          } 
        });
      } else {
        await route.fulfill({ 
          json: { 
            videos: mockVideoEntries, 
            lastRefresh: new Date().toISOString(),
            totalCount: mockVideoEntries.length 
          } 
        });
      }
    });

    await page.route('**/api/channels*', async (route) => {
      const url = route.request().url();
      if (url.includes('empty=true')) {
        await route.fulfill({ json: [] });
      } else {
        await route.fulfill({ json: mockChannels });
      }
    });

    await page.route('**/api/config/smtp*', async (route) => {
      await route.fulfill({ json: mockSMTPConfig });
    });

    await page.route('**/api/newsletter/run*', async (route) => {
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
    await page.route('**/api/videos/**/watch', async (route) => {
      await route.fulfill({ json: { success: true } });
    });
  });

  // Helper function to toggle dark mode
  async function toggleDarkMode(page: Page) {
    await page.evaluate(() => {
      document.documentElement.classList.toggle('dark');
    });
  }

  // Helper function to wait for page load
  async function waitForPageLoad(page: Page) {
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000); // Extra time for animations
  }

  test.describe('Home Page (Videos) Screenshots', () => {
    test('should capture home page with videos - light mode', async ({ page }) => {
      await page.goto('/');
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('home-videos-light.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture home page with videos - dark mode', async ({ page }) => {
      await page.goto('/');
      await toggleDarkMode(page);
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('home-videos-dark.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture home page - empty state', async ({ page }) => {
      await page.goto('/?empty=true');
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('home-empty-light.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture home page with search filter', async ({ page }) => {
      await page.goto('/');
      await waitForPageLoad(page);
      
      // Apply search filter
      await page.fill('input[placeholder*="Search"]', 'React');
      await page.waitForTimeout(500);
      
      await expect(page).toHaveScreenshot('home-search-filtered.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture home page with today filter', async ({ page }) => {
      await page.goto('/');
      await waitForPageLoad(page);
      
      // Toggle today filter
      const todayButton = page.locator('button', { hasText: "Today's Videos" });
      if (await todayButton.isVisible()) {
        await todayButton.click();
        await page.waitForTimeout(500);
      }
      
      await expect(page).toHaveScreenshot('home-today-filtered.png', {
        fullPage: true,
        timeout: 10000,
      });
    });
  });

  test.describe('Subscriptions Page Screenshots', () => {
    test('should capture subscriptions page with channels - light mode', async ({ page }) => {
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('subscriptions-channels-light.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture subscriptions page with channels - dark mode', async ({ page }) => {
      await page.goto('/subscriptions');
      await toggleDarkMode(page);
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('subscriptions-channels-dark.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture subscriptions page - empty state', async ({ page }) => {
      await page.goto('/subscriptions?empty=true');
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('subscriptions-empty-light.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture subscriptions page with import modal', async ({ page }) => {
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      // Open import modal
      const importButton = page.locator('button', { hasText: 'Import Channels' });
      await importButton.click();
      await page.waitForTimeout(500);
      
      await expect(page).toHaveScreenshot('subscriptions-import-modal.png', {
        fullPage: true,
        timeout: 10000,
      });
    });
  });

  test.describe('Notifications/Settings Page Screenshots', () => {
    test('should capture notifications page - light mode', async ({ page }) => {
      await page.goto('/notifications');
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('notifications-light.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture notifications page - dark mode', async ({ page }) => {
      await page.goto('/notifications');
      await toggleDarkMode(page);
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('notifications-dark.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture notifications page with newsletter results', async ({ page }) => {
      await page.goto('/notifications');
      await waitForPageLoad(page);
      
      // Trigger newsletter run to show results
      const runButton = page.locator('button', { hasText: 'Run Newsletter' });
      if (await runButton.isVisible()) {
        await runButton.click();
        await page.waitForTimeout(1000);
      }
      
      await expect(page).toHaveScreenshot('notifications-newsletter-results.png', {
        fullPage: true,
        timeout: 10000,
      });
    });
  });

  test.describe('Responsive Design Screenshots', () => {
    test('should capture mobile view - home page', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/');
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('mobile-home-light.png', {
        fullPage: true,
        timeout: 10000,
      });
    });

    test('should capture tablet view - subscriptions page', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 }); // iPad
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      await expect(page).toHaveScreenshot('tablet-subscriptions-light.png', {
        fullPage: true,
        timeout: 10000,
      });
    });
  });
});