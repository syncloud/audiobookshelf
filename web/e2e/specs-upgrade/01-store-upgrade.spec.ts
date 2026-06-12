import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { installStoreVersion, upgradeToBuild } from '../helpers/device'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

test('the published store version upgrades to this build keeping session and data', async ({ page }, info) => {
  test.setTimeout(1_800_000)

  installStoreVersion()
  await loginViaSyncloud(page, baseURL, username, password)
  await shoot(page, info, 'store-version')

  upgradeToBuild()
  await page.goto(baseURL)
  await expect(page.getByRole('toolbar', { name: 'Appbar' })).toBeVisible({ timeout: 120_000 })
  await shoot(page, info, 'after-upgrade')
})
