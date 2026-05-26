import { test, expect } from '@playwright/test'

test.describe('Merge - camino fail', () => {
  test('rechaza dos PDFs corruptos con 500 y muestra alert en la UI', async ({ page }) => {
    let mergeStatus = 0
    let mergeContentType = ''
    let dialogMessage = ''

    // Persistent handler: dismiss every alert immediately so a native dialog
    // never blocks the page thread (an unhandled alert is what stalled this test).
    page.on('dialog', async (dialog) => {
      dialogMessage = dialog.message()
      await dialog.dismiss()
    })

    page.on('response', (res) => {
      if (res.url().includes('/api/v1/merge') && res.request().method() === 'POST') {
        mergeStatus = res.status()
        mergeContentType = res.headers()['content-type'] || ''
      }
    })

    await page.goto('/gopdfsuit/#/merge')

    await expect(
      page.getByRole('heading', { name: /PDF Merge Tool/i }),
    ).toBeVisible()

    const fileInput = page.locator('input[accept=".pdf"]')
    await fileInput.setInputFiles([
      {
        name: 'corrupted-1.pdf',
        mimeType: 'application/pdf',
        buffer: Buffer.from('not a real pdf #1 - just garbage bytes'),
      },
      {
        name: 'corrupted-2.pdf',
        mimeType: 'application/pdf',
        buffer: Buffer.from('still not a pdf #2 - more garbage'),
      },
    ])

    await expect(page.getByText('corrupted-1.pdf')).toBeVisible()
    await expect(page.getByText('corrupted-2.pdf')).toBeVisible()

    const mergeBtn = page.getByRole('button', { name: /Merge PDFs/i })
    await expect(mergeBtn).toBeEnabled()
    await mergeBtn.click()

    // Backend rejects the garbage upload with a JSON 500.
    await expect.poll(() => mergeStatus, { timeout: 30_000 }).toBe(500)
    expect(mergeContentType).toContain('application/json')

    // The UI surfaces the error via an alert dialog.
    await expect.poll(() => dialogMessage, { timeout: 10_000 }).toContain('Error merging PDFs')
    expect(dialogMessage).toContain('no valid PDF files to merge')

    // No merged result should be rendered, and the button stays usable.
    await expect(page.getByTitle('Merged PDF')).toHaveCount(0)
    await expect(
      page.getByRole('button', { name: /Download Merged PDF/i }),
    ).toHaveCount(0)
    await expect(mergeBtn).toBeEnabled()
  })
})
