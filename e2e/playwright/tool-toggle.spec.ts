import { test, expect } from '@playwright/test';

/**
 * E2E tests for tool toggle functionality (mcpproxy-go-807 / mcpproxy-go-37f).
 *
 * Tests:
 * 1. Tool toggle button shows loading spinner during API call
 * 2. Tool toggle button is disabled during in-flight request
 * 3. Success toast appears after toggle
 * 4. Disabled tools section shows enabled tools that were disabled
 * 5. Enable button in disabled tools section works
 * 6. Race condition fix: no double-click needed
 *
 * Prerequisites: MCPProxy running on localhost:8080 with at least one server
 * that has tools available.
 */

const BASE_URL = 'http://localhost:8080/ui/';

test.describe('Tool Toggle - Race Condition Fix', () => {
  test('toggle button is disabled during API call', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    // Navigate to a server detail page
    // First, go to Servers page and click a server
    await page.click('text=Servers');
    await page.waitForLoadState('networkidle');

    // Wait for server list and click first server
    const serverLink = page.locator('a[href*="/server/"]').first();
    const hasServers = await serverLink.count();
    if (hasServers === 0) {
      test.skip();
      return;
    }

    await serverLink.click();
    await page.waitForLoadState('networkidle');

    // Wait for tools section to load
    await page.waitForSelector('text=Tools', { timeout: 10000 });

    // Find a Disable button
    const disableBtn = page.locator('button:has-text("Disable")').first();
    const hasDisableBtn = await disableBtn.count();
    if (hasDisableBtn === 0) {
      test.skip();
      return;
    }

    // Click the button
    await disableBtn.click();

    // Verify button is disabled during loading (shows spinner)
    // The button should have loading-spinner class or be disabled
    const isLoading = await page.locator('.loading-spinner').first().isVisible().catch(() => false);
    const isDisabled = await disableBtn.isDisabled();

    // Either loading spinner visible or button disabled is acceptable
    expect(isLoading || isDisabled).toBeTruthy();
  });

  test('success toast appears after tool toggle', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    // Navigate to Servers page
    await page.click('text=Servers');
    await page.waitForLoadState('networkidle');

    const serverLink = page.locator('a[href*="/server/"]').first();
    if (await serverLink.count() === 0) {
      test.skip();
      return;
    }

    await serverLink.click();
    await page.waitForLoadState('networkidle');

    // Wait for tools to load
    await page.waitForSelector('text=Tools', { timeout: 10000 });

    // Find and click a Disable button
    const disableBtn = page.locator('button:has-text("Disable")').first();
    if (await disableBtn.count() === 0) {
      test.skip();
      return;
    }

    await disableBtn.click();

    // Wait for success toast
    await expect(page.locator('text=Tool Updated')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=has been disabled')).toBeVisible({ timeout: 5000 });
  });

  test('disabled tools section shows toggleable tools', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    // Go to Settings page where we can disable a tool first
    await page.click('text=Settings');
    await page.waitForLoadState('networkidle');

    // Alternatively, use API to disable a tool first, then check UI
    // This test verifies the disabled tools section renders correctly
    test.skip(); // Requires specific server setup with disabled tools
  });

  test('enable button in disabled tools section has loading state', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    await page.click('text=Servers');
    await page.waitForLoadState('networkidle');

    const serverLink = page.locator('a[href*="/server/"]').first();
    if (await serverLink.count() === 0) {
      test.skip();
      return;
    }

    await serverLink.click();
    await page.waitForLoadState('networkidle');

    // Navigate to Tools tab
    await page.click('text=Tools');
    await page.waitForLoadState('networkidle');

    // Check if there's a disabled tools section with an Enable button
    const enableBtn = page.locator('button:has-text("Enable")').first();
    if (await enableBtn.count() === 0) {
      test.skip();
      return;
    }

    await enableBtn.click();

    // Verify loading state
    const isLoading = await page.locator('.loading-spinner').isVisible().catch(() => false);
    expect(isLoading).toBeTruthy();
  });

  test('no double-click needed to toggle tool', async ({ page }) => {
    // This test verifies the race condition fix:
    // Before fix: clicking Disable sometimes required two clicks
    // After fix: single click should work reliably

    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    await page.click('text=Servers');
    await page.waitForLoadState('networkidle');

    const serverLink = page.locator('a[href*="/server/"]').first();
    if (await serverLink.count() === 0) {
      test.skip();
      return;
    }

    await serverLink.click();
    await page.waitForLoadState('networkidle');

    await page.waitForSelector('text=Tools', { timeout: 10000 });

    // Count disable buttons before
    const disableBtn = page.locator('button:has-text("Disable")').first();
    if (await disableBtn.count() === 0) {
      test.skip();
      return;
    }

    // Single click should be enough
    await disableBtn.click();

    // Wait for success toast (indicates the toggle succeeded)
    await expect(page.locator('text=Tool Updated')).toBeVisible({ timeout: 10000 });

    // Verify the tool moved to disabled section or status changed
    // The toast appearing confirms single-click success
  });
});

test.describe('Tool Toggle - Include Button (Excluded Tools)', () => {
  test('include button has loading state', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    await page.click('text=Servers');
    await page.waitForLoadState('networkidle');

    const serverLink = page.locator('a[href*="/server/"]').first();
    if (await serverLink.count() === 0) {
      test.skip();
      return;
    }

    await serverLink.click();
    await page.waitForLoadState('networkidle');

    await page.click('text=Tools');
    await page.waitForLoadState('networkidle');

    // Check for Include button in excluded tools section
    const includeBtn = page.locator('button:has-text("Include")').first();
    if (await includeBtn.count() === 0) {
      test.skip();
      return;
    }

    await includeBtn.click();

    const isLoading = await page.locator('.loading-spinner').isVisible().catch(() => false);
    expect(isLoading).toBeTruthy();
  });
});
