import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import Viewer from '../../frontend/src/pages/Viewer'
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

describe('Viewer Component (Frontend - TC 05)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.alert = vi.fn()
  })

  it('TC 05: Frontend solicita la generación de portada (Pass)', async () => {
    // Mock the API to return a successful blob
    const mockBlob = new Blob(['pdf-template-content'], { type: 'application/pdf' })
    makeAuthenticatedRequest.mockResolvedValueOnce({
      ok: true,
      blob: () => Promise.resolve(mockBlob)
    })

    render(<Viewer />)

    // Find textarea and enter JSON
    const textArea = screen.getByPlaceholderText(/Enter or paste your JSON template here/i)
    fireEvent.change(textArea, { target: { value: '{"title": "Test Cover"}' } })

    // Click Generate PDF button
    const generateBtn = screen.getByRole('button', { name: /Generate PDF/i })
    expect(generateBtn).not.toBeDisabled()
    fireEvent.click(generateBtn)

    // Verify API call was made to /api/v1/generate/template-pdf
    await waitFor(() => {
      expect(makeAuthenticatedRequest).toHaveBeenCalledWith(
        '/api/v1/generate/template-pdf',
        expect.objectContaining({
          method: 'POST',
          body: '{"title":"Test Cover"}'
        }),
        expect.any(Function)
      )
    })

    // Verify createObjectURL was called
    await waitFor(() => {
      expect(window.URL.createObjectURL).toHaveBeenCalled()
    })

    // Verify the iframe preview and download button appear
    expect(await screen.findByTitle('PDF Preview')).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /Download PDF/i })).toBeInTheDocument()
  })
})
