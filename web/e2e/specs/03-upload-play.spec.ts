import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { getToken, createLibrary, uploadBook, waitForFirstItemId } from '../helpers/abs'
import { ssh } from '../helpers/ssh'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const samplePath = fileURLToPath(new URL('../fixtures/sample.mp3', import.meta.url))

test('admin uploads a book via the UI and plays it', async ({ page }, info) => {
  const suffix = `${info.project.name}-${info.retry}`
  const libraryName = `Books ${suffix}`
  const libraryPath = `/data/audiobookshelf/library-${suffix}`
  ssh(`mkdir -p ${libraryPath} && chown -R audiobookshelf:audiobookshelf ${libraryPath}`)

  await loginViaSyncloud(page, baseURL, username, password)
  await shoot(page, info, 'logged-in')

  const token = await getToken(page)
  const libraryId = await createLibrary(page, token, libraryName, libraryPath)

  await page.goto(`${baseURL}/audiobookshelf/library/${libraryId}`)
  await expect(page.getByRole('toolbar', { name: 'Appbar' })).toBeVisible({ timeout: 45_000 })

  await uploadBook(page, libraryName, samplePath)
  await shoot(page, info, 'uploaded')

  const itemId = await waitForFirstItemId(page, token, libraryId)
  await page.goto(`${baseURL}/audiobookshelf/item/${itemId}`)

  const playButton = page.getByRole('button', { name: 'Play', exact: true }).first()
  await expect(playButton).toBeVisible({ timeout: 30_000 })
  await shoot(page, info, 'item')
  await playButton.click()

  await expect(page.locator('#mediaPlayerContainer')).toBeVisible({ timeout: 30_000 })
  await shoot(page, info, 'playing')
})
