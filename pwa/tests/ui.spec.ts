import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const threeDbPath = path.resolve(__dirname, '../../pwsafe/test_dbs/three.dat');

test.describe('UI Improvements', () => {

    test('should create new database', async ({ page }) => {
        // Mock window.showOpenFilePicker not needed for creation
        // But we might need to mock if we proceed to open it?
        // Actually creation doesn't use picker.

        await page.goto('/');

        // Open Menu
        await page.locator('.hamburger').click();
        await page.getByText('Create New DB').click();

        // Check for header
        await expect(page.locator('h2')).toHaveText('Create New Database');

        // Enter password
        await page.getByPlaceholder('New Password').fill('newpass123');
        await page.getByRole('button', { name: 'Create' }).click();

        // Verify Dashboard loads (Unlock success)
        // Should have empty tree
        await expect(page.locator('.sidebar')).toBeVisible();
        // search input check
        await expect(page.getByPlaceholder('Search...')).toBeVisible();
    });

    test('should have functioning dashboard menu', async ({ page }) => {
        // Load DB first
        const buffer = fs.readFileSync(threeDbPath);
        const data = [...buffer];

        await page.addInitScript((fileData) => {
            (window as any).showOpenFilePicker = async () => {
                const blob = new Blob([new Uint8Array(fileData)], { type: 'application/octet-stream' });
                const file = new File([blob], 'three.dat');
                return [{
                    getFile: async () => file,
                    name: 'three.dat'
                }];
            };
        }, data);

        await page.goto('/');
        await page.getByText('Open Database File').click();
        await page.getByPlaceholder('Password').fill('three3#;');
        await page.getByRole('button', { name: 'Unlock' }).click();

        // 1. Autofocus Check
        // Need to wait slightly for timeout
        await page.waitForTimeout(200);
        await expect(page.getByPlaceholder('Search...')).toBeFocused();

        // 2. DB Info Check
        await page.locator('.hamburger').click();

        // Handle alert
        page.once('dialog', dialog => {
            expect(dialog.message()).toContain('DB Info:');
            dialog.dismiss();
        });
        await page.getByText('DB Info').click();

        // 3. Save DB Check
        await page.locator('.hamburger').click();
        page.once('dialog', dialog => {
            expect(dialog.message()).toContain('not yet implemented');
            dialog.dismiss();
        });
        await page.getByText('Save DB').click();

        // 4. Close DB Check
        await page.locator('.hamburger').click();
        await page.getByText('Close DB').click();
        await expect(page.locator('.start-page h1')).toHaveText('Password Safe');
    });

    test('should handle mobile layout', async ({ page }) => {
        // Set mobile viewport
        await page.setViewportSize({ width: 375, height: 667 });

        // Load DB
        const buffer = fs.readFileSync(threeDbPath);
        const data = [...buffer];

        await page.addInitScript((fileData) => {
            (window as any).showOpenFilePicker = async () => {
                const blob = new Blob([new Uint8Array(fileData)], { type: 'application/octet-stream' });
                const file = new File([blob], 'three.dat');
                return [{
                    getFile: async () => file,
                    name: 'three.dat'
                }];
            };
        }, data);

        await page.goto('/');
        await page.getByText('Open Database File').click();
        await page.getByPlaceholder('Password').fill('three3#;');
        await page.getByRole('button', { name: 'Unlock' }).click();

        // Select an item
        const firstItem = page.locator('.tree li').first();
        await firstItem.click();

        // Verify Main Content covers sidebar
        // In mobile, sidebar width is 100%, and main-content is fixed overlay.
        // We can check if main-content has class mobile-open
        const mainContent = page.locator('.main-content');
        await expect(mainContent).toHaveClass(/mobile-open/);
        await expect(mainContent).toBeVisible();

        // Verify Close button exists and works
        const closeBtn = page.locator('.close-details');
        await expect(closeBtn).toBeVisible();
        await closeBtn.click();

        // Verify returns to list
        await expect(mainContent).not.toHaveClass(/mobile-open/);
    });

});
