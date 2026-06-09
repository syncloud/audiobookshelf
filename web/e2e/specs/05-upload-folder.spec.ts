import { test, expect } from '@playwright/test'
import { readdirSync } from 'fs'
import { join } from 'path'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { uploadFolderFiles } from '../helpers/abs'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const libraryName = 'Books'
const audiobookDir = process.env.PLAYWRIGHT_AUDIOBOOK_DIR || ''

test('admin uploads a multi-file audiobook folder', async ({ page }, info) => {
  test.skip(!audiobookDir, 'PLAYWRIGHT_AUDIOBOOK_DIR not set')
  test.setTimeout(600_000)

  const files = readdirSync(audiobookDir)
    .filter((f) => f.endsWith('.mp3'))
    .map((f) => join(audiobookDir, f))
  expect(files.length).toBeGreaterThan(1)

  await loginViaSyncloud(page, baseURL, username, password)
  await uploadFolderFiles(page, libraryName, files)
  await shoot(page, info, 'folder-uploaded')
})
