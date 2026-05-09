import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const threeDbPath = path.resolve(__dirname, '../../pwsafe/test_dbs/three.dat');

// three.dat contents (password: "three3#;"):
//   "three entry 1"  group="group1"   user="three1_user"  url="http://group1.com"
//   "three entry 2"  group="group2"   user="three2_user"
//   "three entry 3"  group="group 3"  user="three3_user"  url="https://group3.com"

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

test.describe('Search', () => {

    test('names only is the default', async ({ page }) => {
        await openThreeDb(page);
        await expect(page.getByLabel('Names only')).toBeChecked();
    });

    test('names-only search finds title matches', async ({ page }) => {
        await openThreeDb(page);
        await page.getByPlaceholder(/Search/).fill('entry 1');
        await expect(page.locator('.tree li')).toHaveCount(1);
        await expect(page.locator('.tree li')).toContainText('three entry 1');
    });

    test('names-only search finds group matches', async ({ page }) => {
        await openThreeDb(page);
        await page.getByPlaceholder(/Search/).fill('group2');
        await expect(page.locator('.tree li')).toHaveCount(1);
        await expect(page.locator('.tree li')).toContainText('three entry 2');
    });

    test('names-only search does not find username', async ({ page }) => {
        await openThreeDb(page);
        await page.getByPlaceholder(/Search/).fill('three1_user');
        await expect(page.locator('.tree li')).toHaveCount(0);
    });

    test('names-only search does not find URL', async ({ page }) => {
        await openThreeDb(page);
        await page.getByPlaceholder(/Search/).fill('group1.com');
        await expect(page.locator('.tree li')).toHaveCount(0);
    });

    test('AND search narrows results', async ({ page }) => {
        await openThreeDb(page);
        await page.getByPlaceholder(/Search/).fill('entry group1');
        await expect(page.locator('.tree li')).toHaveCount(1);
        await expect(page.locator('.tree li')).toContainText('three entry 1');
    });

    test('search details finds username', async ({ page }) => {
        await openThreeDb(page);
        await page.getByLabel('Names only').uncheck();
        await page.getByPlaceholder(/Search/).fill('three1_user');
        await expect(page.locator('.tree li')).toHaveCount(1);
        await expect(page.locator('.tree li')).toContainText('three entry 1');
    });

    test('search details finds URL', async ({ page }) => {
        await openThreeDb(page);
        await page.getByLabel('Names only').uncheck();
        await page.getByPlaceholder(/Search/).fill('group3.com');
        await expect(page.locator('.tree li')).toHaveCount(1);
        await expect(page.locator('.tree li')).toContainText('three entry 3');
    });

    test('search details AND search spans multiple fields', async ({ page }) => {
        await openThreeDb(page);
        await page.getByLabel('Names only').uncheck();
        // "entry" matches title, "group1.com" matches URL — both must be true
        await page.getByPlaceholder(/Search/).fill('entry group1.com');
        await expect(page.locator('.tree li')).toHaveCount(1);
        await expect(page.locator('.tree li')).toContainText('three entry 1');
    });

    test('unmatched term returns empty list', async ({ page }) => {
        await openThreeDb(page);
        await page.getByPlaceholder(/Search/).fill('zzznotfound');
        await expect(page.locator('.tree li')).toHaveCount(0);
    });

});
