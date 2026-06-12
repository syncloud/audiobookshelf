import { test } from '@playwright/test'
import { installStoreVersion, upgradeToBuild, inspect } from '../helpers/device'

test('diagnose store install + upgrade restart/db behavior', async ({ page }, info) => {
  test.setTimeout(1_800_000)

  installStoreVersion()
  console.log(inspect('after store install'))
  await page.waitForTimeout(60_000)
  console.log(inspect('store install +60s'))

  upgradeToBuild()
  console.log(inspect('after upgrade'))
  await page.waitForTimeout(60_000)
  console.log(inspect('upgrade +60s'))
})
