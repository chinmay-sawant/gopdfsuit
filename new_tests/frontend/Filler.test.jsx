import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import Filler from '../../frontend/src/pages/Filler'
import { makeAuthenticatedRequest } from '../../frontend/src/utils/apiConfig'

// Mock apiConfig
vi.mock('../../frontend/src/utils/apiConfig', () => ({
  makeAuthenticatedRequest: vi.fn()
}))

// Mock AuthContext
vi.mock('../../frontend/src/contexts/AuthContext', () => ({
  useAuth: () => ({
    getAuthHeaders: vi.fn(),
    triggerLogin: vi.fn()
  })
}))

// Mock BackgroundAnimation (to avoid canvas/animation issues in tests)
vi.mock('../../frontend/src/components/BackgroundAnimation', () => ({
  default: () => <div data-testid="bg-animation" />
}))

describe('Filler Component (Frontend - TC 02 y TC 03)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.alert = vi.fn()
  })

  it('TC 02: Frontend intenta llenar el AcroForm usando mocks (Pass)', async () => {
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    // Mock the API to return a successful blob
    const mockBlob = new Blob(['pdf-content'], { type: 'application/pdf' })
    makeAuthenticatedRequest.mockResolvedValueOnce({
      ok: true,
      blob: () => Promise.resolve(mockBlob)
    })

    render(<Filler />)

    const pdfInput = document.querySelector('input[accept=".pdf"]')
    const xfdfInput = document.querySelector('input[accept=".xfdf,.xml"]')

    const pdfFile = new File(['pdf data'], 'form.pdf', { type: 'application/pdf' })
    fireEvent.change(pdfInput, { target: { files: [pdfFile] } })

    const xfdfFile = new File(['<xfdf/>'], 'data.xfdf', { type: 'application/xml' })
    fireEvent.change(xfdfInput, { target: { files: [xfdfFile] } })

    // Wait for UI to update labels
    expect(await screen.findByText('form.pdf')).toBeInTheDocument()
    expect(await screen.findByText('data.xfdf')).toBeInTheDocument()

    // Click the Fill button
    const fillButton = screen.getByRole('button', { name: /Fill PDF Form/i })
    expect(fillButton).not.toBeDisabled()
    fireEvent.click(fillButton)

    // Verify API call was made
    await waitFor(() => {
      expect(makeAuthenticatedRequest).toHaveBeenCalledWith(
        '/api/v1/fill',
        expect.objectContaining({ method: 'POST' }),
        expect.any(Function)
      )
    })

    // Verify that createObjectURL was called
    await waitFor(() => {
      expect(window.URL.createObjectURL).toHaveBeenCalled()
    })

    // Verify Download button appears
    expect(await screen.findByRole('button', { name: /Download Filled PDF/i })).toBeInTheDocument()

    expect(clickSpy).toHaveBeenCalled()
    clickSpy.mockRestore()
  })

  it('TC 03: Frontend falla al llenar el AcroForm por error del servidor (Fail)', async () => {
    // Mock the API to throw an error
    makeAuthenticatedRequest.mockRejectedValueOnce(new Error('Internal Server Error HTTP 500'))

    render(<Filler />)

    const pdfInput = document.querySelector('input[accept=".pdf"]')
    const xfdfInput = document.querySelector('input[accept=".xfdf,.xml"]')

    const pdfFile = new File(['pdf data'], 'form.pdf', { type: 'application/pdf' })
    const xfdfFile = new File(['<xfdf/>'], 'data.xfdf', { type: 'application/xml' })

    fireEvent.change(pdfInput, { target: { files: [pdfFile] } })
    fireEvent.change(xfdfInput, { target: { files: [xfdfFile] } })

    const fillButton = screen.getByRole('button', { name: /Fill PDF Form/i })
    
    await waitFor(() => expect(fillButton).not.toBeDisabled())
    
    fireEvent.click(fillButton)

    // Verify alert is called
    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith(expect.stringContaining('Error filling PDF: Internal Server Error HTTP 500'))
    })
  })
})
