import { test, expect } from '@playwright/test';

test.describe('Database Write Operations', () => {
    test('Create DB, Add, Update, Delete Record', async ({ page }) => {
        page.on('console', msg => console.log('PAGE LOG:', msg.text()));
        page.on('pageerror', exception => console.log(`PAGE ERROR: "${exception}"`));

        // 1. Create New Database
        await page.goto('/');
        await page.click('button[aria-label="Menu"]');
        await page.click('text=Create New DB');
        await page.fill('input[placeholder="New Password"]', 'password123');
        await page.click('button:has-text("Create")');

        // Check if opened (Dashboard visible)
        await expect(page.locator('.sidebar')).toBeVisible();

        // 2. Add New Record
        await page.click('button[aria-label="Menu"]');
        await page.click('text=New Record');

        // Fill record details
        await page.fill('input[placeholder="Title"]', 'Test Record');
        await page.fill('input[placeholder="Group"]', 'Test Group');
        await page.fill('input[placeholder="Username"]', 'testuser');
        await page.fill('input[placeholder="Password"]', 'secret123');
        await page.fill('textarea[placeholder="Notes"]', 'Some notes');

        await page.click('text=Save Record');

        // Verify it appears in the tree
        await expect(page.locator('.tree')).toContainText('Test Group');
        await expect(page.locator('.tree')).toContainText('Test Record');

        // 3. Update Record
        await page.click('text=Test Record');
        await page.fill('input[placeholder="Title"]', 'Updated Record');
        await page.click('text=Save Record');

        // Verify tree updates
        await expect(page.locator('.tree')).toContainText('Updated Record');
        await expect(page.locator('.tree')).not.toContainText('Test Record'); // Old title gone

        // 4. Delete Record
        await page.on('dialog', dialog => dialog.accept()); // Handle confirmation dialog
        await page.click('text=Delete Record');

        // Verify gone
        await expect(page.locator('.tree')).not.toContainText('Updated Record');

        // 5. Save DB (Just verify it doesn't crash, difficult to test file system API in headless properly without mocking)
        // We can just verify the button is there.
        await page.click('button[aria-label="Menu"]');
        await expect(page.locator('text=Save DB')).toBeVisible();
    });
});
