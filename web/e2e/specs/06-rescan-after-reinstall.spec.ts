import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { uploadBook } from '../helpers/abs'
import { ssh, scpTo } from '../helpers/ssh'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const libraryName = 'Books'
const bookFile = process.env.PLAYWRIGHT_RESCAN_BOOK_FILE || ''
const snap = process.env.PLAYWRIGHT_SNAP || ''

test('a reinstall rescans the existing library without re-uploading', async ({ page }, info) => {
  test.setTimeout(900_000)

  await loginViaSyncloud(page, baseURL, username, password)
  await uploadBook(page, libraryName, bookFile)
  await shoot(page, info, 'uploaded')

  scpTo(snap, '/tmp/audiobookshelf.snap')
  ssh('snap remove audiobookshelf')
  ssh('snap install --dangerous /tmp/audiobookshelf.snap')

  await expect(async () => {
    await page.goto(baseURL)
    await expect(page.getByRole('link', { name: /login with syncloud/i })).toBeVisible({ timeout: 10_000 })
  }).toPass({ timeout: 600_000 })

  await loginViaSyncloud(page, baseURL, username, password)
  await shoot(page, info, 'after-reinstall')

  const card = page.locator('[id^="book-card-"]').filter({ hasText: 'Rescan Book' }).first()
  await expect(card).toBeVisible({ timeout: 120_000 })
  await shoot(page, info, 'rescanned')
})
