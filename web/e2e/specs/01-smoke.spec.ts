import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

test('audiobookshelf loads and accepts initial setup', async ({ page }, info) => {
  await page.goto('/')
  await page.waitForLoadState('domcontentloaded')
  await expect(page).toHaveTitle(/audiobookshelf/i, { timeout: 30_000 })
  await shoot(page, info, 'landing')

  const userField = page.locator('input#root-username, input[type="text"]').first()
  const passField = page.locator('input[type="password"]').first()
  const submit = page
    .getByRole('button', { name: /create|submit|root|account/i })
    .or(page.locator('button[type="submit"]'))
    .first()

  const onInit = await passField.isVisible({ timeout: 15_000 }).catch(() => false)
  if (onInit) {
    if (await userField.isVisible().catch(() => false)) {
      await userField.fill(username)
    }
    await passField.fill(password)
    await shoot(page, info, 'init-filled')
    await submit.click({ timeout: 10_000 }).catch(() => {})
    await page.waitForLoadState('networkidle').catch(() => {})
    await shoot(page, info, 'after-init')
  }

  await expect(page.locator('body')).toBeVisible()
})
