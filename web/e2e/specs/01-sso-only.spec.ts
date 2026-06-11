import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

test('audiobookshelf shows SSO-only login with no password form', async ({ page }, info) => {
  await page.goto('/')
  await page.waitForLoadState('domcontentloaded')
  await expect(page).toHaveTitle(/audiobookshelf/i, { timeout: 30_000 })

  const ssoButton = page.getByRole('link', { name: /login with syncloud/i })
  await expect(ssoButton).toBeVisible({ timeout: 60_000 })
  await expect(page.locator('input[type="password"]')).toBeHidden()
  await shoot(page, info, 'login')
})
