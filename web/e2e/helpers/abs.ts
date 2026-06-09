import { Page, expect } from '@playwright/test'

export async function getToken(page: Page): Promise<string> {
  const token = await page.evaluate(() => localStorage.getItem('token'))
  if (!token) throw new Error('no auth token in localStorage after login')
  return token
}

export async function uploadBook(page: Page, libraryName: string, fileName: string, buffer: Buffer) {
  await page.getByRole('button', { name: 'Upload Media' }).click()

  const content = page.locator('#app-content')
  await content.getByRole('button', { name: /^Library/ }).click()
  const libraryOption = content.getByRole('menuitem', { name: libraryName })
  await expect(libraryOption).toBeVisible({ timeout: 30_000 })
  await libraryOption.click()

  await page.locator('input[type="file"]').first().setInputFiles({ name: fileName, mimeType: 'audio/mpeg', buffer })

  const uploadButton = page.getByRole('button', { name: 'Upload', exact: true })
  await expect(uploadButton).toBeEnabled({ timeout: 30_000 })
  await uploadButton.click()

  // Fail loudly if the upload errored instead of relying on a book that may already exist.
  await expect(page.getByText(/Successfully Uploaded/i)).toBeVisible({ timeout: 60_000 })
}

export async function uploadFolderFiles(page: Page, libraryName: string, filePaths: string[]) {
  await page.getByRole('button', { name: 'Upload Media' }).click()

  const content = page.locator('#app-content')
  await content.getByRole('button', { name: /^Library/ }).click()
  const libraryOption = content.getByRole('menuitem', { name: libraryName })
  await expect(libraryOption).toBeVisible({ timeout: 30_000 })
  await libraryOption.click()

  await page.locator('input[webkitdirectory]').setInputFiles(filePaths)

  const uploadButton = page.getByRole('button', { name: 'Upload', exact: true })
  await expect(uploadButton).toBeEnabled({ timeout: 60_000 })
  await uploadButton.click()

  await expect(page.getByText(/Successfully Uploaded/i)).toBeVisible({ timeout: 300_000 })
}
