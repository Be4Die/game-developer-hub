import { test, expect } from '@playwright/test';

test.describe('Frontend Baseline Smoke Tests', () => {
  test('Unauthenticated user visiting root redirects to /login', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    await page.goto('/');
    await expect(page).toHaveURL(/.*\/login/);
    await expect(page.locator('input[type="email"]')).toBeVisible();

    // Filter out expected network errors when backend is offline
    const fatalErrors = consoleErrors.filter(
      (e) => !e.includes('ERR_CONNECTION_REFUSED') && !e.includes('500') && !e.includes('Network Error')
    );
    expect(fatalErrors).toEqual([]);
  });

  test('Public documentation route loads and displays content', async ({ page }) => {
    await page.goto('/docs');
    await expect(page).toHaveURL(/.*\/docs\/getting-started\/overview/);
    await expect(page.locator('body')).toBeVisible();
  });

  test('Rules route redirects to rules catalog', async ({ page }) => {
    await page.goto('/rules');
    await expect(page).toHaveURL(/.*\/docs\/rules\/rules-catalog/);
  });
});
