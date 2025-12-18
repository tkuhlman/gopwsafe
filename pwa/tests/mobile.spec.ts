
import { test, expect } from '@playwright/test';

test.use({
    viewport: { width: 375, height: 667 }, // iPhone SE dimensions
});

test('Mobile UI: Check layout and copy buttons', async ({ page }) => {
    // Go to the app
    await page.goto('/');

    // Create a new DB to get to the dashboard
    // Open menu first
    await page.getByLabel('Menu').click();
    await page.getByRole('button', { name: 'Create New DB' }).click();

    // Enter password for new DB
    await page.getByPlaceholder('New Password', { exact: true }).fill('password');
    await page.getByPlaceholder('New Password', { exact: true }).press('Enter');

    // Wait for dashboard to load
    // We look for the search input which is always there
    await expect(page.getByPlaceholder('Search...')).toBeVisible();

    // Create a new record to see fields
    // Open menu again
    await page.getByLabel('Menu').click();
    await page.getByRole('button', { name: 'New Record' }).click();

    // Check if horizontal scrollbar is present (it shouldn't be)
    const scrollWidth = await page.evaluate(() => document.body.scrollWidth);
    const clientWidth = await page.evaluate(() => document.body.clientWidth);
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1); // +1 buffer for rounding

    // Check for Copy icons
    // Based on the code: button with title="Copy Username" and "Copy Password"
    const copyUserBtn = page.getByTitle('Copy Username');
    const copyPassBtn = page.getByTitle('Copy Password');

    await expect(copyUserBtn).toBeVisible();
    await expect(copyPassBtn).toBeVisible();

    // Check that input fields are visible and not crushed
    const usernameInput = page.getByPlaceholder('Username');
    await expect(usernameInput).toBeVisible();
    const box = await usernameInput.boundingBox();
    // Ensure it has some significant width (e.g., > 100px on mobile) to prove flex works
    expect(box.width).toBeGreaterThan(150);

    // Verify "Show" button is also visible and password input shares space
    // For a new record, the password is shown by default, so button says "Hide"
    const showBtn = page.getByRole('button', { name: 'Hide' });
    await expect(showBtn).toBeVisible();

    const passwordInput = page.getByPlaceholder('Password');
    const passBox = await passwordInput.boundingBox();
    expect(passBox.width).toBeGreaterThan(100);
});
