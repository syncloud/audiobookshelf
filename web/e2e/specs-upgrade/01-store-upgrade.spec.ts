import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { installStoreVersion, upgradeToBuild } from '../helpers/device'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

async function login (page) {
  await expect(async () => {
    await loginViaSyncloud(page, baseURL, username, password)
  }).toPass({ timeout: 600_000, intervals: [5_000] })
}

test('the published store version upgrades to this build and stays usable', async ({ page }, info) => {
  test.setTimeout(1_800_000)

  installStoreVersion()
  await login(page)
  await shoot(page, info, 'store-version')

  upgradeToBuild()
  await login(page)
  await shoot(page, info, 'after-upgrade')
})
