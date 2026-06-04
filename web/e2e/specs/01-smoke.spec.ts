import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

test('audiobookshelf loads and reaches login or initial setup', async ({ page }, info) => {
  await page.goto('/')
  await page.waitForLoadState('domcontentloaded')
  await expect(page).toHaveTitle(/audiobookshelf/i, { timeout: 30_000 })

  const formField = page.locator('input[type="password"], input[type="text"], input').first()
  await expect(formField).toBeVisible({ timeout: 60_000 })
  await shoot(page, info, 'login')
})
