import { Page, expect } from '@playwright/test'

export async function getToken(page: Page): Promise<string> {
  const token = await page.evaluate(() => localStorage.getItem('token'))
  if (!token) throw new Error('no auth token in localStorage after login')
  return token
}

export async function createLibrary(page: Page, token: string, name: string, folderPath: string): Promise<string> {
  const result = await page.evaluate(async ({ name, folderPath, token }) => {
    const r = await fetch('/audiobookshelf/api/libraries', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ name, mediaType: 'book', folders: [{ fullPath: folderPath }] })
    })
    return { ok: r.ok, status: r.status, body: await r.text() }
  }, { name, folderPath, token })
  if (!result.ok) throw new Error(`create library failed: ${result.status} ${result.body}`)
  return JSON.parse(result.body).id
}

export async function uploadBook(page: Page, libraryName: string, filePath: string) {
  await page.getByRole('button', { name: 'Upload Media' }).click()

  const content = page.locator('#app-content')
  await content.getByRole('button', { name: /^Library/ }).click()
  await content.getByRole('menuitem', { name: libraryName }).click()

  await page.locator('input[type="file"]').first().setInputFiles(filePath)

  const uploadButton = page.getByRole('button', { name: 'Upload', exact: true })
  await expect(uploadButton).toBeEnabled({ timeout: 30_000 })
  await uploadButton.click()
}

export async function waitForFirstItemId(page: Page, token: string, libraryId: string): Promise<string> {
  let itemId = ''
  await expect(async () => {
    const results = await page.evaluate(async ({ libraryId, token }) => {
      const r = await fetch(`/audiobookshelf/api/libraries/${libraryId}/items`, { headers: { Authorization: `Bearer ${token}` } })
      if (!r.ok) return null
      return (await r.json()).results
    }, { libraryId, token })
    expect(results && results.length).toBeGreaterThan(0)
    itemId = results[0].id
  }).toPass({ timeout: 90_000, intervals: [2000] })
  return itemId
}
