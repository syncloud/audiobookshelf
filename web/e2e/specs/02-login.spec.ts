import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

test('logs in via Syncloud SSO and lands in the app', async ({ page }, info) => {
  await loginViaSyncloud(page, baseURL, username, password)
  await expect(page).not.toHaveURL(/\/login/)
  await shoot(page, info, 'after-login')
})
