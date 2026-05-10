import { test, expect } from '@playwright/test';

// Tests use a freshly-created in-memory DB with known records so the
// most-frequent suggestion is deterministic.

async function setupDb(page) {
    await page.addInitScript(() => {
        (window as any).showSaveFilePicker = async () => ({
            createWritable: async () => ({ write: async () => {}, close: async () => {} }),
            name: 'test.psafe3',
        });
    });
    await page.goto('/');
    await page.click('button[aria-label="Menu"]');
    await page.click('text=Create New DB');
    await page.fill('input[placeholder="New Password"]', 'password123');
    await page.click('button:has-text("Create")');
    await expect(page.locator('.sidebar')).toBeVisible();
}

async function addRecord(page, title: string, group: string, username: string) {
    await page.click('button[aria-label="Menu"]');
    await page.click('text=New Record');
    await page.fill('input[placeholder="Title"]', title);
    await page.fill('input[placeholder="Group"]', group);
    await page.fill('input[aria-label="Username"]', username);
    await page.click('text=Save Record');
    await expect(page.locator('.tree')).toContainText(title);
}

test.describe('Autocomplete', () => {

    test('ghost text appears for group prefix match', async ({ page }) => {
        await setupDb(page);
        await addRecord(page, 'Record 1', 'Engineering', 'alice');
        await addRecord(page, 'Record 2', 'Engineering', 'alice');

        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');
        await page.fill('input[placeholder="Group"]', 'Eng');

        await expect(page.locator('.ghost-suffix').first()).toHaveText('ineering');
    });

    test('Tab completes group suggestion', async ({ page }) => {
        await setupDb(page);
        await addRecord(page, 'Record 1', 'Engineering', 'alice');

        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');
        const groupInput = page.locator('input[placeholder="Group"]');
        await groupInput.fill('Eng');
        await expect(page.locator('.ghost-suffix').first()).toBeVisible();
        await groupInput.press('Tab');
        await expect(groupInput).toHaveValue('Engineering');
    });

    test('ghost text appears for username prefix match', async ({ page }) => {
        await setupDb(page);
        await addRecord(page, 'Record 1', 'Engineering', 'alice');

        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');
        await page.fill('input[aria-label="Username"]', 'ali');

        await expect(page.locator('.ghost-suffix').last()).toHaveText('ce');
    });

    test('Tab completes username suggestion', async ({ page }) => {
        await setupDb(page);
        await addRecord(page, 'Record 1', 'Engineering', 'alice');

        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');
        const usernameInput = page.locator('input[aria-label="Username"]');
        await usernameInput.fill('ali');
        await expect(page.locator('.ghost-suffix').last()).toBeVisible();
        await usernameInput.press('Tab');
        await expect(usernameInput).toHaveValue('alice');
    });

    test('Escape dismisses ghost suggestion', async ({ page }) => {
        await setupDb(page);
        await addRecord(page, 'Record 1', 'Engineering', 'alice');

        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');
        const groupInput = page.locator('input[placeholder="Group"]');
        await groupInput.fill('Eng');
        await expect(page.locator('.ghost-suffix').first()).toBeVisible();
        await groupInput.press('Escape');
        await expect(page.locator('.ghost-overlay')).toHaveCount(0);
    });

    test('no ghost text when prefix has no match', async ({ page }) => {
        await setupDb(page);
        await addRecord(page, 'Record 1', 'Engineering', 'alice');

        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');
        await page.fill('input[placeholder="Group"]', 'xyz');
        await expect(page.locator('.ghost-overlay')).toHaveCount(0);
    });

    test('most frequent value is suggested when multiple match', async ({ page }) => {
        await setupDb(page);
        await addRecord(page, 'Record 1', 'Engineering', 'alice');
        await addRecord(page, 'Record 2', 'Engineering', 'alice');
        await addRecord(page, 'Record 3', 'Entertainment', 'alice');

        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');
        await page.fill('input[placeholder="Group"]', 'En');

        // Engineering appears twice, Entertainment once — Engineering wins
        await expect(page.locator('.ghost-suffix').first()).toHaveText('gineering');
    });

});
