// Thin route wrapper: the shared page lives in HtmlConvertPage.jsx
// ({mode: 'image'}). Kept so /htmltoimage keeps working unchanged.
import HtmlConvertPage from './HtmlConvertPage'

const HtmlToImage = () => <HtmlConvertPage mode="image" />

export default HtmlToImage
