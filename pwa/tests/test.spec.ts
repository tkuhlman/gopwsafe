import { test, expect } from '@playwright/test';

test('has title', async ({ page }) => {
    await page.goto('/');

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Password Safe/);
});

test('has open db button', async ({ page }) => {
    await page.goto('/');

    // Check for the "Open DB" button or input
    // Adjust selector based on actual implementation
    // Looking at StartPage.svelte might be useful if this fails, but guessing standard elements for now.
    // Or checking text content.
    await expect(page.getByRole('button', { name: /Open/i })).toBeVisible();
    // or 
    // await expect(page.getByText('Open Password Safe File')).toBeVisible();
});
