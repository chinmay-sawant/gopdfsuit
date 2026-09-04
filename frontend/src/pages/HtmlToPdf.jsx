// Thin route wrapper: the shared page lives in HtmlConvertPage.jsx
// ({mode: 'pdf'}). Kept so /htmltopdf keeps working unchanged.
import HtmlConvertPage from './HtmlConvertPage'

const HtmlToPdf = () => <HtmlConvertPage mode="pdf" />

export default HtmlToPdf
