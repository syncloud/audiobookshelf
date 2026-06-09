import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { readFileSync } from 'fs'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { uploadBook } from '../helpers/abs'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const libraryName = 'Books'
const sampleBuffer = readFileSync(fileURLToPath(new URL('../fixtures/sample.mp3', import.meta.url)))

test('admin uploads a book to the default library and plays it', async ({ page }, info) => {
  await loginViaSyncloud(page, baseURL, username, password)
  await expect(page.getByRole('toolbar', { name: 'Appbar' })).toBeVisible({ timeout: 45_000 })
  await shoot(page, info, 'logged-in')

  // The default 'Books' library created during install must be selectable in the upload dropdown.
  // Unique filename per project/attempt so desktop and mobile (same device) don't collide.
  const fileName = `sample-${info.project.name}-${info.retry}.mp3`
  await uploadBook(page, libraryName, fileName, sampleBuffer)
  await shoot(page, info, 'uploaded')

  // Back to the bookshelf, wait for the scanned book card to appear, open it.
  await page.locator('#appbar').getByRole('link').first().click()
  const card = page.locator('#book-card-0').first()
  await expect(card).toBeVisible({ timeout: 90_000 })
  await expect(card).toContainText('Test Book')
  await shoot(page, info, 'bookshelf')
  await card.click()

  const playButton = page.getByRole('button', { name: 'Play', exact: true }).first()
  await expect(playButton).toBeVisible({ timeout: 30_000 })
  await shoot(page, info, 'item')
  await playButton.click()

  await expect(page.locator('#mediaPlayerContainer')).toBeVisible({ timeout: 30_000 })
  await shoot(page, info, 'playing')
})
