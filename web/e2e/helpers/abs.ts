import { Page, expect } from '@playwright/test'

export async function getToken(page: Page): Promise<string> {
  const token = await page.evaluate(() => localStorage.getItem('token'))
  if (!token) throw new Error('no auth token in localStorage after login')
  return token
}

export async function createLibrary(page: Page, baseURL: string, token: string, name: string, folderPath: string): Promise<string> {
  const res = await page.request.post(`${baseURL}/audiobookshelf/api/libraries`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { name, mediaType: 'book', folders: [{ fullPath: folderPath }] }
  })
  if (!res.ok()) throw new Error(`create library failed: ${res.status()} ${await res.text()}`)
  return (await res.json()).id
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

export async function waitForFirstItemId(page: Page, baseURL: string, token: string, libraryId: string): Promise<string> {
  let itemId = ''
  await expect(async () => {
    const res = await page.request.get(`${baseURL}/audiobookshelf/api/libraries/${libraryId}/items`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(res.ok()).toBeTruthy()
    const results = (await res.json()).results
    expect(results.length).toBeGreaterThan(0)
    itemId = results[0].id
  }).toPass({ timeout: 90_000, intervals: [2000] })
  return itemId
}
