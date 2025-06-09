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

  // Helper function to toggle dark mode
  async function toggleDarkMode(page: Page) {
    await page.evaluate(() => {
      // Add dark class for Tailwind dark mode
      document.documentElement.classList.add('dark');
      
      // Override CSS custom properties to force dark mode
      const root = document.documentElement;
      
      // Set the root variables that are used in @media (prefers-color-scheme: dark)
      root.style.setProperty('--background', '#0a0a0a');
      root.style.setProperty('--foreground', '#ededed');
      
      // Also set the theme variables that Tailwind uses
      root.style.setProperty('--color-background', '#0a0a0a');
      root.style.setProperty('--color-foreground', '#ededed');
      
      // Force body styles as well to ensure immediate effect
      document.body.style.background = '#0a0a0a';
      document.body.style.color = '#ededed';
      
      // Add data attribute to indicate dark mode for better debugging
      root.setAttribute('data-theme', 'dark');
      
      // Force all elements with dark mode classes to apply their styles
      // This ensures Tailwind's dark mode classes take effect immediately
      const allElements = document.querySelectorAll('*');
      allElements.forEach(el => {
        const element = el as HTMLElement;
        if (element.className && element.className.includes('dark:')) {
          // Force recompute styles by touching a style property
          const display = getComputedStyle(element).display;
          element.style.display = 'none';
          element.offsetHeight; // Trigger reflow
          element.style.display = display;
        }
      });
      
      // Additional forced styles for common card elements to ensure dark mode
      const cards = document.querySelectorAll('.bg-white, [class*="bg-white"]');
      cards.forEach(card => {
        const element = card as HTMLElement;
        if (element.className.includes('dark:bg-gray-800')) {
          element.style.backgroundColor = '#1f2937'; // gray-800
        }
      });
      
      const borders = document.querySelectorAll('[class*="border-gray-200"]');
      borders.forEach(border => {
        const element = border as HTMLElement;
        if (element.className.includes('dark:border-gray-700')) {
          element.style.borderColor = '#374151'; // gray-700
        }
      });
      
      // Force text colors for dark mode
      const grayTexts = document.querySelectorAll('[class*="text-gray-600"]');
      grayTexts.forEach(text => {
        const element = text as HTMLElement;
        if (element.className.includes('dark:text-gray-400')) {
          element.style.color = '#9ca3af'; // gray-400
        }
      });
      
      // Force a style recalculation by accessing offsetHeight
      document.body.offsetHeight;
    });
    
    // Wait for dark mode styles to be applied and any transitions to complete
    await page.waitForTimeout(2000);
    
    // Verify dark mode is applied by checking for the dark class and actual styles
    const isDarkModeActive = await page.evaluate(() => {
      const hasDarkClass = document.documentElement.classList.contains('dark');
      
      // Also check if dark mode styles are actually being applied
      // by checking the computed background color of a card element
      const cardElement = document.querySelector('.bg-white');
      const actualBg = cardElement ? getComputedStyle(cardElement).backgroundColor : '';
      
      return {
        hasDarkClass,
        actualBackgroundColor: actualBg,
        rootBackground: getComputedStyle(document.documentElement).getPropertyValue('--background').trim()
      };
    });
    
    console.log('Dark mode verification:', isDarkModeActive);
    
    if (!isDarkModeActive.hasDarkClass) {
      throw new Error('Dark mode class was not properly applied');
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

  test.describe('Home Page (Videos) Screenshots', () => {
    test('should capture home page with videos - light mode', async ({ page }) => {
      await page.goto('/');
      await waitForPageLoad(page);
      
      // Wait for videos to load by checking for video cards
      await page.waitForSelector('[data-testid="video-card"], .video-card, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'home-videos-light.png');
    });

    test('should capture home page with videos - dark mode', async ({ page }) => {
      await page.goto('/');
      await toggleDarkMode(page);
      await waitForPageLoad(page);
      
      // Wait for videos to load
      await page.waitForSelector('[data-testid="video-card"], .video-card, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'home-videos-dark.png');
    });

    test('should capture home page - empty state', async ({ page }) => {
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

  test.describe('Subscriptions Page Screenshots', () => {
    test('should capture subscriptions page with channels - light mode', async ({ page }) => {
      await page.goto('/subscriptions');
      await waitForPageLoad(page);
      
      // Wait for the page title to ensure content is loaded
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'subscriptions-channels-light.png');
    });

    test('should capture subscriptions page with channels - dark mode', async ({ page }) => {
      await page.goto('/subscriptions');
      await toggleDarkMode(page);
      await waitForPageLoad(page);
      
      await page.waitForSelector('h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'subscriptions-channels-dark.png');
    });

    test('should capture subscriptions page - empty state', async ({ page }) => {
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

  test.describe('Notifications/Settings Page Screenshots', () => {
    test('should capture notifications page - light mode', async ({ page }) => {
      await page.goto('/notifications');
      await waitForPageLoad(page);
      
      // Wait for forms to load
      await page.waitForSelector('form, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'notifications-light.png');
    });

    test('should capture notifications page - dark mode', async ({ page }) => {
      await page.goto('/notifications');
      await toggleDarkMode(page);
      await waitForPageLoad(page);
      
      await page.waitForSelector('form, h1', { timeout: 10000 });
      
      await captureScreenshot(page, 'notifications-dark.png');
    });
  });

  test.describe('Responsive Design Screenshots', () => {
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
});