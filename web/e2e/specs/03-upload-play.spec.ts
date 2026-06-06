import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { createLibraryViaApi, uploadBook } from '../helpers/abs'
import { ssh } from '../helpers/ssh'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const libraryName = 'Test Library'
const libraryPath = '/data/audiobookshelf/library'
const samplePath = fileURLToPath(new URL('../fixtures/sample.mp3', import.meta.url))

test('admin uploads a book via the UI and plays it', async ({ page }, info) => {
  ssh(`mkdir -p ${libraryPath} && chown -R audiobookshelf:audiobookshelf ${libraryPath}`)

  await loginViaSyncloud(page, baseURL, username, password)
  await shoot(page, info, 'logged-in')

  await createLibraryViaApi(page, baseURL, libraryName, libraryPath)
  await page.goto(baseURL)
  await expect(page.getByRole('toolbar', { name: 'Appbar' })).toBeVisible({ timeout: 45_000 })

  await uploadBook(page, libraryName, samplePath)
  await shoot(page, info, 'uploaded')

  await page.goto(baseURL)
  const item = page.getByText('Test Book').first()
  await expect(item).toBeVisible({ timeout: 60_000 })
  await item.click()

  const playButton = page.getByRole('button', { name: 'Play', exact: true }).first()
  await expect(playButton).toBeVisible({ timeout: 30_000 })
  await shoot(page, info, 'item')
  await playButton.click()

  await expect(page.locator('#mediaPlayerContainer')).toBeVisible({ timeout: 30_000 })
  await shoot(page, info, 'playing')
})
