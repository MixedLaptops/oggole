import { test, expect } from '@playwright/test';

test('application serves the main page', async ({ page }) => {
  await page.goto('/');

  // Check page loads successfully
  await expect(page).toHaveURL('http://localhost:8080/');
});

test('login page is accessible', async ({ page }) => {
  await page.goto('/login');

  // Check for username and password fields
  await expect(page.locator('input[name="username"]')).toBeVisible();
  await expect(page.locator('input[name="password"]')).toBeVisible();
});

test('register page is accessible', async ({ page }) => {
  await page.goto('/register');

  // Check for registration form fields
  await expect(page.locator('input[name="username"]')).toBeVisible();
  await expect(page.locator('input[name="email"]')).toBeVisible();
});
