import {devices, expect, test} from '@playwright/test';

test('guest can search, open an offer, and use the mobile filter drawer', async ({page}) => {
  await page.goto('/ar', {waitUntil: 'domcontentloaded'});
  await expect(page.locator('.splash-screen')).toBeHidden({timeout: 15000});
  const search = page.getByRole('combobox', {name: 'بحث'});
  await expect(search).toBeVisible();
  await search.fill('ازياء');
  await expect(page.getByRole('listbox', {name: 'اقتراحات البحث'})).toBeVisible();
  const firstSuggestion = page.getByRole('option').first();
  await search.press('ArrowDown');
  await expect(firstSuggestion).toHaveClass(/is-active/);
  await search.press('Enter');
  await expect(search).toHaveValue(/الأزياء/);
  await expect(page.getByRole('link', {name: /تخفيضات الصيف/}).first()).toBeVisible();
  await page.locator('a[href="/ar/offers/offer-001"]').last().click();
  await expect(page).toHaveURL(/\/ar\/offers\/offer-001/);
  await expect(page.getByRole('heading', {name: /تخفيضات الصيف/})).toBeVisible();
  await expect(page.getByRole('button', {name: /مشاركة/})).toBeVisible();
});

test('mobile filter drawer opens from the header and search stays on the page', async ({browser}) => {
  const context = await browser.newContext({...devices['iPhone 13'], locale: 'ar-DZ'});
  const page = await context.newPage();
  await page.goto('/ar/offers', {waitUntil: 'domcontentloaded'});
  await expect(page.locator('.splash-screen')).toBeHidden({timeout: 15000});
  await expect(page.getByRole('combobox', {name: 'بحث'})).toBeVisible();
  await page.getByRole('button', {name: 'فتح البحث والفلاتر'}).click();
  await expect(page.getByRole('complementary', {name: 'الفلاتر'})).toHaveClass(/is-open/);
  await expect(page.getByRole('combobox', {name: 'بحث'})).toBeVisible();
  await context.close();
});
