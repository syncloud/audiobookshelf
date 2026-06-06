import { Page, expect } from '@playwright/test'

export async function createLibraryViaApi(page: Page, baseURL: string, name: string, folderPath: string) {
  const token = await page.evaluate(() => localStorage.getItem('token'))
  if (!token) throw new Error('no auth token in localStorage after login')

  const res = await page.request.post(`${baseURL}/audiobookshelf/api/libraries`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { name, mediaType: 'book', folders: [{ fullPath: folderPath }] }
  })
  if (!res.ok()) throw new Error(`create library failed: ${res.status()} ${await res.text()}`)
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
