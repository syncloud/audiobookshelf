import { test, expect } from '@playwright/test'
import { loginViaSyncloud } from '../helpers/auth'
import { getToken } from '../helpers/abs'

const domain = process.env.PLAYWRIGHT_DOMAIN || 'bookworm.com'
const baseURL = `https://audiobookshelf.${domain}`
const username = process.env.PLAYWRIGHT_USER || 'user'
const password = process.env.PLAYWRIGHT_PASSWORD || 'Password1'

const storageRoot = '/data/audiobookshelf'

async function browse(page, token: string, query: string): Promise<string[]> {
  return page.evaluate(async ({ query, token }) => {
    const r = await fetch(`/audiobookshelf/api/filesystem${query}`, { headers: { Authorization: `Bearer ${token}` } })
    if (!r.ok) throw new Error(`filesystem api ${r.status}`)
    return (await r.json()).directories.map((d: any) => d.path)
  }, { query, token })
}

test('folder browser is limited to the app storage root', async ({ page }) => {
  await loginViaSyncloud(page, baseURL, username, password)
  const token = await getToken(page)

  for (const query of ['', '?path=/', '?path=/usr']) {
    const paths = await browse(page, token, query)
    expect(paths.length).toBeGreaterThan(0)
    for (const p of paths) {
      expect(p.startsWith(storageRoot)).toBeTruthy()
    }
  }
})
