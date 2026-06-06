import { Page, expect } from '@playwright/test'

export async function loginViaSyncloud(page: Page, baseURL: string, username: string, password: string) {
  const appHost = new URL(baseURL).host

  page.on('framenavigated', (frame) => {
    if (frame === page.mainFrame()) console.log('[nav]', frame.url())
  })

  await page.goto(baseURL)

  const ssoButton = page.getByRole('link', { name: /login with syncloud/i })
  await expect(ssoButton).toBeVisible({ timeout: 60_000 })
  await ssoButton.click()

  const usernameField = page.locator('#username-textfield')
  await expect(usernameField).toBeVisible({ timeout: 30_000 })
  await usernameField.fill(username)
  await page.locator('#password-textfield').fill(password)
  await page.locator('#sign-in-button').click()

  try {
    await page.waitForURL((url) => url.host === appHost && !url.pathname.startsWith('/login') && !url.pathname.startsWith('/auth/'), { timeout: 45_000 })
  } catch (e) {
    console.log('[final url]', page.url())
    throw e
  }
  await expect(page).toHaveTitle(/audiobookshelf/i, { timeout: 30_000 })
  await expect(page.getByRole('link', { name: /login with syncloud/i })).toHaveCount(0)
}
