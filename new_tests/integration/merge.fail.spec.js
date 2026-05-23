import { test, expect } from '@playwright/test'

test.describe('Merge - camino fail', () => {
  test('rechaza dos PDFs corruptos con 500 y muestra alert en la UI', async ({ page }) => {
    test.setTimeout(90_000)

    let mergeStatus = 0
    let mergeContentType = ''

    page.on('response', (res) => {
      if (
        res.url().includes('/api/v1/merge') &&
        res.request().method() === 'POST'
      ) {
        mergeStatus = res.status()
        mergeContentType = res.headers()['content-type'] || ''
      }
    })

    await page.goto('/gopdfsuit/#/merge')

    await expect(
      page.getByRole('heading', { name: /PDF Merge Tool/i }),
    ).toBeVisible()

    const dialogPromise = page.waitForEvent('dialog')

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

    const dialog = await dialogPromise

    expect(mergeStatus).toBe(500)
    expect(mergeContentType).toContain('application/json')

    expect(dialog.type()).toBe('alert')
    expect(dialog.message()).toContain('Error merging PDFs')
    expect(dialog.message()).toContain('no valid PDF files to merge')
    await dialog.dismiss()

    await expect(page.getByTitle('Merged PDF')).toHaveCount(0)
    await expect(
      page.getByRole('button', { name: /Download Merged PDF/i }),
    ).toHaveCount(0)

    await expect(mergeBtn).toBeEnabled()
  })
})
