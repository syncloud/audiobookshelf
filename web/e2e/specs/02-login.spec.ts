import { test } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'

const domain = process.env.PLAYWRIGHT_DOMAIN
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER
const password = process.env.PLAYWRIGHT_PASSWORD

test('logs in via Syncloud SSO and lands in the app', async ({ page }, info) => {
  await loginViaSyncloud(page, baseURL, username, password)
  await shoot(page, info, 'after-login')
})
