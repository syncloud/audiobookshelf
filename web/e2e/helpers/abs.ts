import { Page, expect } from '@playwright/test'

export async function createLibraryViaApi(page: Page, baseURL: string, name: string, folderPath: string) {
  const token = await page.evaluate(() => localStorage.getItem('token'))
  console.log('[diag] token present:', !!token)
  if (!token) throw new Error('no auth token in localStorage after login')
  const headers = { Authorization: `Bearer ${token}` }

  const me = await page.request.get(`${baseURL}/audiobookshelf/api/me`, { headers })
  console.log('[diag] me:', me.status(), await me.text())

  const res = await page.request.post(`${baseURL}/audiobookshelf/api/libraries`, {
    headers,
    data: { name, mediaType: 'book', folders: [{ fullPath: folderPath }] }
  })
  console.log('[diag] create library:', res.status(), await res.text())
  if (!res.ok()) throw new Error(`create library failed: ${res.status()}`)

  const libs = await page.request.get(`${baseURL}/audiobookshelf/api/libraries`, { headers })
  console.log('[diag] libraries after create:', libs.status(), await libs.text())
}

export async function uploadBook(page: Page, libraryName: string, filePath: string) {
  await page.getByRole('button', { name: 'Upload Media' }).click()

  await page.getByRole('button', { name: /^Library/ }).click()
  await page.getByRole('menuitem', { name: libraryName }).click()

  await page.locator('input[type="file"]').first().setInputFiles(filePath)

  const uploadButton = page.getByRole('button', { name: 'Upload', exact: true })
  await expect(uploadButton).toBeEnabled({ timeout: 30_000 })
  await uploadButton.click()
}
