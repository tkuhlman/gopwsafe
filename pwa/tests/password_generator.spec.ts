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
    await expect(page.getByPlaceholder('Search...')).toBeVisible();
}

test.describe('Password options panel', () => {

    test('gear icon toggles the panel open and closed', async ({ page }) => {
        await openThreeDb(page);
        await page.locator('.tree li').first().click();
        await expect(page.locator('.record-details')).toBeVisible();

        const panel = page.locator('.pwgen-panel');
        await expect(panel).not.toBeVisible();

        await page.getByTitle('Generator options').click();
        await expect(panel).toBeVisible();

        await page.getByTitle('Generator options').click();
        await expect(panel).not.toBeVisible();
    });

    test('Generate button produces a new password', async ({ page }) => {
        await openThreeDb(page);
        await page.locator('.tree li').first().click();

        const passwordInput = page.locator('.record-details').getByPlaceholder('Password');
        const before = await passwordInput.inputValue();

        await page.getByTitle('Generator options').click();
        await page.getByRole('button', { name: 'Generate' }).click();

        const after = await passwordInput.inputValue();
        expect(after).not.toBe(before);
        expect(after.length).toBeGreaterThan(0);
    });

});
