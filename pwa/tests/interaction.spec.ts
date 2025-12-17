import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Paths to test databases
const simpleDbPath = path.resolve(__dirname, '../../pwsafe/test_dbs/simple.dat');
const threeDbPath = path.resolve(__dirname, '../../pwsafe/test_dbs/three.dat');

test.describe('Database Interaction', () => {

    test('should fail to open DB with wrong password', async ({ page }) => {
        // Read the file buffer
        const buffer = fs.readFileSync(simpleDbPath);
        const data = [...buffer]; // Convert to array for passing to browser

        // Mock showOpenFilePicker
        await page.addInitScript((fileData) => {
            (window as any).showOpenFilePicker = async () => {
                const blob = new Blob([new Uint8Array(fileData)], { type: 'application/octet-stream' });
                const file = new File([blob], 'simple.dat');
                return [{
                    getFile: async () => file,
                    name: 'simple.dat'
                }];
            };
        }, data);

        await page.goto('/');

        // Click Open Database File
        await page.getByText('Open Database File').click();

        // Expect password prompt
        await expect(page.getByText('Unlock simple.dat')).toBeVisible();

        // Enter wrong password
        await page.getByPlaceholder('Password').fill('wrongpassword');
        await page.getByRole('button', { name: 'Unlock' }).click();

        // Expect error
        await expect(page.locator('.error')).toContainText('Failed to unlock');
    });

    test('should open DB with correct password and search', async ({ page }) => {
        // Read the file buffer for three.dat
        // three.dat has password "three3#;"
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

        // Enter correct password
        await page.getByPlaceholder('Password').fill('three3#;');
        await page.getByRole('button', { name: 'Unlock' }).click();

        // Expect StartPage to disappear and Main interface to appear
        await expect(page.getByPlaceholder('Search...')).toBeVisible();

        // Check for groups in sidebar
        await expect(page.locator('.sidebar details summary').first()).toBeVisible();

        // Pick the first item in the list
        const firstItem = page.locator('.tree li').first();

        // Ensure there is at least one item
        await expect(firstItem).toBeVisible();

        const title = await firstItem.textContent();
        await firstItem.click();

        // Verify details view appears
        await expect(page.locator('.record-details h2')).toHaveText(title?.trim() || '');

        // Verify password masking
        const passwordInput = page.locator('.password-row input');
        await expect(passwordInput).toHaveAttribute('type', 'password');

        // Toggle password visibility
        await page.getByRole('button', { name: 'Show' }).click();
        await expect(passwordInput).toHaveAttribute('type', 'text');
        await page.getByRole('button', { name: 'Hide' }).click();
        await expect(passwordInput).toHaveAttribute('type', 'password');

        // Test Search
        const searchInput = page.getByPlaceholder('Search...');
        await searchInput.fill(title?.trim() || 'xyz');

        // Should still see the item
        await expect(page.locator('.tree li', { hasText: title?.trim() })).toBeVisible();

        // Search for something non-existent
        await searchInput.fill('NonExistentItem12345');
        await expect(page.locator('.tree li')).toBeHidden(); // Or check count is 0
    });

    test('should copy username and password to clipboard', async ({ page, context }) => {
        // Grant clipboard permissions
        await context.grantPermissions(['clipboard-read', 'clipboard-write']);

        // Load three.dat
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

        // Select first item
        const firstItem = page.locator('.tree li').first();
        await expect(firstItem).toBeVisible();
        await firstItem.click();

        // Test Copy Username Button
        const copyUserBtn = page.getByTitle('Copy Username');
        await expect(copyUserBtn).toBeVisible();
        await copyUserBtn.click();
        await expect(page.getByText('Copied!').first()).toBeVisible();
        // Wait for it to disappear
        await expect(page.getByText('Copied!').first()).toBeHidden({ timeout: 3000 });

        // Test Copy Password Button
        const copyPassBtn = page.getByTitle('Copy Password');
        await expect(copyPassBtn).toBeVisible();
        await copyPassBtn.click();
        await expect(page.getByText('Copied!').last()).toBeVisible(); // Might be the same if transient, but we have two locations
        // Wait for it to disappear
        await expect(page.getByText('Copied!').last()).toBeHidden({ timeout: 3000 });

        // Test Shortcuts
        // Ctrl+U
        await page.keyboard.press('Control+u');
        await expect(page.getByText('Copied!').first()).toBeVisible();

        // Reset by waiting or just proceed (it disappears in 2s)
        await page.waitForTimeout(2100);

        // Ctrl+P
        await page.keyboard.press('Control+p');
        await expect(page.getByText('Copied!').last()).toBeVisible();
    });
});
