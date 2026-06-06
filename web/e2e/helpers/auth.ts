import { Page, expect } from '@playwright/test'

export async function loginViaSyncloud(page: Page, baseURL: string, username: string, password: string) {
  await page.goto(baseURL)

  const ssoButton = page.getByRole('link', { name: /login with syncloud/i })
  await expect(ssoButton).toBeVisible({ timeout: 60_000 })
  await ssoButton.click()

  const usernameField = page.locator('#username-textfield')
  await expect(usernameField).toBeVisible({ timeout: 30_000 })
  await usernameField.fill(username)
  await page.locator('#password-textfield').fill(password)
  await page.locator('#sign-in-button').click()

  await expect(page.getByRole('toolbar', { name: 'Appbar' })).toBeVisible({ timeout: 45_000 })
}
