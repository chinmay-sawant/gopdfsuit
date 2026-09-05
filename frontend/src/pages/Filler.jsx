import { FileCheck, FileText, RefreshCw } from 'lucide-react'
import { useState } from 'react'
import ConsentBanner from '../components/ConsentBanner'
import FileDropzone from '../components/FileDropzone'
import OperationShell from '../components/OperationShell'
import OpPageShell from '../components/OpPageShell'
import { useAuth } from '../contexts/AuthContext'
import { usePdfOperation } from '../hooks/usePdfOperation'
import { fillPDFSmart, fillViaServer, shouldUseServerWasmTransport } from '../utils/wasmLoader.js'

const serverTransport = shouldUseServerWasmTransport()

export default function Filler() {
  const [pdfFile, setPdfFile] = useState(null)
  const [xfdfFile, setXfdfFile] = useState(null)
  const { getAuthHeaders, triggerLogin } = useAuth()
  const [fallbackOffer, setFallbackOffer] = useState(null)
  const { isLoading, resultUrl, run, runLocal, download } = usePdfOperation({
    onAuthRequired: triggerLogin,
    onError: (message) => alert(`Error filling PDF: ${message}`),
  })
  // Preview stays hidden when files are only selected. It appears while
  // filling or after the filled result exists.
  const hasPreview = Boolean(isLoading || resultUrl)

  const fillPDF = async () => {
    if (!pdfFile || !xfdfFile) return

    if (serverTransport) {
      const formData = new FormData()
      formData.append('pdf', pdfFile)
      formData.append('xfdf', xfdfFile)
      await run({
        endpoint: '/api/v1/fill',
        body: formData,
        getAuthHeaders,
        filename: `filled-${pdfFile.name}`,
      })
      return
    }

    setFallbackOffer(null)
    const pdfBytes = new Uint8Array(await pdfFile.arrayBuffer())
    const xfdfBytes = new Uint8Array(await xfdfFile.arrayBuffer())
    const pdfName = pdfFile.name
    let wasmMessage = ''
    const url = await runLocal(() => fillPDFSmart(pdfBytes, xfdfBytes, pdfName, { getAuthHeaders }), {
      autoDownload: false,
      onError: (message) => { wasmMessage = message },
    })
    if (!url && getAuthHeaders) {
      setFallbackOffer({ pdfBytes, xfdfBytes, pdfName, message: wasmMessage })
    }
  }

  const fillViaServerConsent = async () => {
    if (!fallbackOffer || isLoading) return
    const { pdfBytes, xfdfBytes, pdfName } = fallbackOffer
    setFallbackOffer(null)
    await runLocal(() => fillViaServer(pdfBytes, xfdfBytes, pdfName, getAuthHeaders), {
      filename: `filled-${pdfName}`,
    })
  }

  return (
    <OpPageShell
      title="Fill a PDF form"
      className={`filler-page tool-wide${hasPreview ? '' : ' tool-single-page'}`}
      icon={<FileCheck aria-hidden="true" size={31} />}
      description={serverTransport
        ? 'This configuration sends the PDF and XFDF file to the fill endpoint.'
        : 'Choose an AcroForm PDF and its XFDF data. The tool explains before it sends either file to the service.'}
    >
      <ConsentBanner
        actionLabel="Upload to server and fill"
        isLoading={isLoading}
        offer={fallbackOffer}
        onConsent={fillViaServerConsent}
        onDismiss={() => setFallbackOffer(null)}
      />
      <div className="container-full filler-fit">
      <div className={`filler-grid${hasPreview ? '' : ' tool-single'}`}>
        <section className="glass-card filler-input-card">
          <h2 className="workspace-card-title">Choose the files</h2>
          <div className="filler-dropzone-stack">
            <FileDropzone
              accept=".pdf,application/pdf"
              disabled={isLoading}
              onFiles={(files) => {
                const selected = files.find((file) => file.type === 'application/pdf')
                if (selected) setPdfFile(selected)
              }}
              subtitle={pdfFile ? pdfFile.name : 'Select the PDF that contains AcroForm fields'}
              title="Choose the PDF"
            />
            <FileDropzone
              accept=".xfdf,.xml,application/xml,text/xml"
              disabled={isLoading}
              onFiles={(files) => {
                const [selected] = files
                if (selected) setXfdfFile(selected)
              }}
              subtitle={xfdfFile ? xfdfFile.name : 'Select the XFDF or XML field data'}
              title="Choose the form data"
            />
          </div>
          <button
            className="btn-glow"
            disabled={isLoading || !pdfFile || !xfdfFile}
            onClick={fillPDF}
            style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', padding: '1rem 2rem' }}
            type="button"
          >
            {isLoading ? <RefreshCw aria-hidden="true" className="spin" size={18} /> : <FileCheck aria-hidden="true" size={18} />}
            Fill PDF form
          </button>
          <p className="filler-hint">The PDF needs AcroForm fields. The XFDF file names the matching fields and values. Text fields, checkboxes, and radio buttons are supported.</p>
        </section>

        {hasPreview && (
        <OperationShell
          resultUrl={resultUrl}
          title="Filled PDF"
          icon={<FileText size={18} />}
          emptyTitle="Your filled PDF will appear here"
          emptySubtitle="Choose both files, then run the fill operation."
          onDownload={() => download(`filled-${pdfFile?.name || 'form.pdf'}`)}
          downloadLabel="Download filled PDF"
        />
        )}
      </div>
      </div>
    </OpPageShell>
  )
}
