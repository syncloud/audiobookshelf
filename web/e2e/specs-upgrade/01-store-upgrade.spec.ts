import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'
import { loginViaSyncloud } from '../helpers/auth'
import { installStoreVersion, upgradeToBuild } from '../helpers/device'

const domain = process.env.PLAYWRIGHT_DOMAIN
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER
const password = process.env.PLAYWRIGHT_PASSWORD

test('the published store version upgrades to this build keeping session and data', async ({ page }, info) => {
  test.setTimeout(1_800_000)

  installStoreVersion()
  await expect(async () => {
    await loginViaSyncloud(page, baseURL, username, password)
  }).toPass({ timeout: 300_000, intervals: [5_000] })
  await shoot(page, info, 'store-version')

  upgradeToBuild()
  await expect(async () => {
    await page.goto(baseURL)
    await expect(page.getByRole('toolbar', { name: 'Appbar' })).toBeVisible({ timeout: 10_000 })
  }).toPass({ timeout: 300_000, intervals: [5_000] })
  await shoot(page, info, 'after-upgrade')
})
