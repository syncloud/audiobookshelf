import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { uploadBook } from '../helpers/abs'

const domain = process.env.PLAYWRIGHT_DOMAIN
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER
const password = process.env.PLAYWRIGHT_PASSWORD

const libraryName = 'Books'
const bookFile = process.env.PLAYWRIGHT_BOOK_FILE || ''

test('admin uploads a book to the default library and plays it', async ({ page }, info) => {
  await loginViaSyncloud(page, baseURL, username, password)
  await expect(page.getByRole('toolbar', { name: 'Appbar' })).toBeVisible({ timeout: 45_000 })
  await shoot(page, info, 'logged-in')

  await uploadBook(page, libraryName, bookFile)
  await shoot(page, info, 'uploaded')

  await page.locator('#appbar').getByRole('link').first().click()
  const card = page.locator('#book-card-0').first()
  await expect(card).toBeVisible({ timeout: 90_000 })
  await expect(card).toContainText('Test Book')
  await shoot(page, info, 'bookshelf')
  await card.click()
  await expect(page.locator('#mediaPlayerContainer')).toBeVisible({ timeout: 30_000 })
  await shoot(page, info, 'playing')
})
