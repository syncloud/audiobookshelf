import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

test('audiobookshelf loads and reaches login or initial setup', async ({ page }, info) => {
  const consoleErrors: string[] = []
  const failed: string[] = []
  page.on('console', (m) => { if (m.type() === 'error') consoleErrors.push(m.text()) })
  page.on('requestfailed', (r) => failed.push(`${r.failure()?.errorText} ${r.url()}`))
  page.on('response', (r) => { if (r.status() >= 400) failed.push(`HTTP ${r.status()} ${r.url()}`) })

  await page.goto('/')
  await page.waitForLoadState('domcontentloaded')
  await expect(page).toHaveTitle(/audiobookshelf/i, { timeout: 30_000 })
  await shoot(page, info, 'landing')

  const formField = page.locator('input[type="password"], input[type="text"], input').first()
  const loaded = await formField.isVisible({ timeout: 60_000 }).catch(() => false)

  console.log('=== console errors ===\n' + consoleErrors.join('\n'))
  console.log('=== failed/4xx-5xx requests ===\n' + failed.join('\n'))
  if (!loaded) {
    console.log('=== page content (first 4000 chars) ===\n' + (await page.content()).slice(0, 4000))
  }
  await shoot(page, info, 'after-load')

  expect(loaded, 'a login/init form input should become visible').toBeTruthy()
})
