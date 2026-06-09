import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { uploadFolder } from '../helpers/abs'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const libraryName = 'Books'
const audiobookDir = process.env.PLAYWRIGHT_AUDIOBOOK_DIR || ''

test('admin uploads a multi-file audiobook folder', async ({ page }, info) => {
  test.skip(!audiobookDir, 'PLAYWRIGHT_AUDIOBOOK_DIR not set')
  test.setTimeout(600_000)

  await loginViaSyncloud(page, baseURL, username, password)
  await uploadFolder(page, libraryName, audiobookDir)
  await shoot(page, info, 'folder-uploaded')
})
