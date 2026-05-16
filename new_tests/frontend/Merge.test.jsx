import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import MergePage from '../../frontend/src/pages/Merge'
import { makeAuthenticatedRequest } from '../../frontend/src/utils/apiConfig'

vi.mock('../../frontend/src/utils/apiConfig', () => ({
  makeAuthenticatedRequest: vi.fn()
}))

vi.mock('../../frontend/src/contexts/AuthContext', () => ({
  useAuth: () => ({
    getAuthHeaders: vi.fn(),
    triggerLogin: vi.fn()
  })
}))

vi.mock('../../frontend/src/components/BackgroundAnimation', () => ({
  default: () => <div data-testid="bg-animation" />
}))

describe('Merge Component (Frontend - TC 07)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.alert = vi.fn()
  })

  it('TC 07: Frontend une la portada y el PDF lleno (Pass)', async () => {
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    // Mock the API to return a successful blob for the merged PDF
    const mockBlob = new Blob(['merged-pdf-content'], { type: 'application/pdf' })
    makeAuthenticatedRequest.mockResolvedValueOnce({
      ok: true,
      blob: () => Promise.resolve(mockBlob)
    })

    render(<MergePage />)

    // Find the file input for PDFs
    const fileInput = document.querySelector('input[accept=".pdf"]')

    // Create mock PDF files
    const coverPdf = new File(['cover data'], 'cover.pdf', { type: 'application/pdf' })
    const filledPdf = new File(['filled form data'], 'filled.pdf', { type: 'application/pdf' })

    // Simulate uploading both files
    fireEvent.change(fileInput, { target: { files: [coverPdf, filledPdf] } })

    // Wait for the files to appear in the selected files list
    expect(await screen.findByText('cover.pdf')).toBeInTheDocument()
    expect(await screen.findByText('filled.pdf')).toBeInTheDocument()

    // Click Merge button
    const mergeBtn = screen.getByRole('button', { name: /Merge PDFs/i })
    expect(mergeBtn).not.toBeDisabled()
    fireEvent.click(mergeBtn)

    // Verify API call was made to /api/v1/merge
    await waitFor(() => {
      expect(makeAuthenticatedRequest).toHaveBeenCalledWith(
        '/api/v1/merge',
        expect.objectContaining({
          method: 'POST',
          body: expect.any(FormData)
        }),
        expect.any(Function)
      )
    })

    // Verify createObjectURL was called
    await waitFor(() => {
      expect(window.URL.createObjectURL).toHaveBeenCalled()
    })

    // Verify the iframe preview and download button appear
    expect(await screen.findByTitle('Merged PDF')).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /Download Merged PDF/i })).toBeInTheDocument()

    expect(clickSpy).toHaveBeenCalled()
    clickSpy.mockRestore()
  })
})
