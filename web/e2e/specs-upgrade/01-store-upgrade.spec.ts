import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { uploadBook } from '../helpers/abs'
import { installStoreVersion, upgradeToBuild } from '../helpers/device'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const libraryName = 'Books'
const bookFile = process.env.PLAYWRIGHT_BOOK_FILE || ''

async function waitForLoginPage (page) {
  await expect(async () => {
    await page.goto(baseURL)
    await expect(page.getByRole('link', { name: /login with syncloud/i })).toBeVisible({ timeout: 10_000 })
  }).toPass({ timeout: 900_000 })
}

test('the published store version upgrades to this build with the library intact', async ({ page }, info) => {
  test.setTimeout(2_400_000)

  installStoreVersion()
  await waitForLoginPage(page)
  await loginViaSyncloud(page, baseURL, username, password)
  await uploadBook(page, libraryName, bookFile)
  await shoot(page, info, 'before-upgrade')

  upgradeToBuild()
  await waitForLoginPage(page)
  await loginViaSyncloud(page, baseURL, username, password)
  await shoot(page, info, 'after-upgrade')

  const card = page.locator('[id^="book-card-"]').filter({ hasText: 'Test Book' }).first()
  await expect(card).toBeVisible({ timeout: 120_000 })
  await shoot(page, info, 'library-intact')
})
