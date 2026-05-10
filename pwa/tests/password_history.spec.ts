import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const threeDbPath = path.resolve(__dirname, '../../pwsafe/test_dbs/three.dat');

// three.dat contents (password: "three3#;"):
//   "three entry 1"  group="group1"   user="three1_user"  pass="three1!@$%^&*()"  url="http://group1.com"
//   "three entry 2"  group="group2"   user="three2_user"  pass="three2_-+=\\|][}{';:"
//   "three entry 3"  group="group 3"  user="three3_user"  pass=",./<>?`~0"        url="https://group3.com"

async function openThreeDb(page) {
    const buffer = fs.readFileSync(threeDbPath);
    const data = [...buffer];
    await page.addInitScript((fileData) => {
        (window as any).showOpenFilePicker = async () => {
            const blob = new Blob([new Uint8Array(fileData)], { type: 'application/octet-stream' });
            const file = new File([blob], 'three.dat');
            return [{
                getFile: async () => file,
                createWritable: async () => ({ write: async () => {}, close: async () => {} }),
                name: 'three.dat',
            }];
        };
    }, data);
    await page.goto('/');
    await page.getByText('Open Database File').click();
    await page.getByPlaceholder('Password').fill('three3#;');
    await page.getByRole('button', { name: 'Unlock' }).click();
    await expect(page.getByPlaceholder(/Search/)).toBeVisible();
}

test.describe('Password history', () => {

    test('previous password appears in history after change', async ({ page }) => {
        await openThreeDb(page);

        // Select first record and note its current password
        await page.locator('.tree li').first().click();
        await page.getByRole('button', { name: 'Show' }).click();
        const originalPassword = await page.locator('#record-password').inputValue();

        // Change the password and save
        await page.locator('#record-password').fill('brand-new-password-xyz');
        await page.getByRole('button', { name: 'Save Record' }).click();

        // Open gear panel
        await page.getByTitle('Password options').click();

        // History toggle should appear
        const historyToggle = page.getByText(/previous password/);
        await expect(historyToggle).toBeVisible();

        // Expand history
        await historyToggle.click();

        // Old password should be shown
        await expect(page.locator('.history-pw')).toContainText(originalPassword);
    });

    test('history copy button copies old password to clipboard', async ({ page, context }) => {
        await context.grantPermissions(['clipboard-read', 'clipboard-write']);
        await openThreeDb(page);

        // Change the password to create a history entry
        await page.locator('.tree li').first().click();
        await page.getByRole('button', { name: 'Show' }).click();
        const originalPassword = await page.locator('#record-password').inputValue();
        await page.locator('#record-password').fill('brand-new-password-xyz');
        await page.getByRole('button', { name: 'Save Record' }).click();

        // Open and expand history
        await page.getByTitle('Password options').click();
        await page.getByText(/previous password/).click();

        // Click the copy button on the history entry
        await page.locator('.history-entry .icon-btn').click();

        const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
        expect(clipboardText).toBe(originalPassword);
    });

});
