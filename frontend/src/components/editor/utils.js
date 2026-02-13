
// Standard margin in points (1 inch = 72 points)
export const DEFAULT_MARGIN = 72

export const parsePageMargins = (marginString) => {
    const defaults = { left: DEFAULT_MARGIN, right: DEFAULT_MARGIN, top: DEFAULT_MARGIN, bottom: DEFAULT_MARGIN }
    if (!marginString || typeof marginString !== 'string') return defaults

    const [left, right, top, bottom] = marginString.split(':').map(v => Number.parseFloat(v))
    return {
        left: Number.isFinite(left) && left >= 0 ? left : defaults.left,
        right: Number.isFinite(right) && right >= 0 ? right : defaults.right,
        top: Number.isFinite(top) && top >= 0 ? top : defaults.top,
        bottom: Number.isFinite(bottom) && bottom >= 0 ? bottom : defaults.bottom
    }
}

export const formatPageMargins = (margins) => {
    const safe = {
        left: Math.max(0, Number.parseFloat(margins?.left ?? DEFAULT_MARGIN) || 0),
        right: Math.max(0, Number.parseFloat(margins?.right ?? DEFAULT_MARGIN) || 0),
        top: Math.max(0, Number.parseFloat(margins?.top ?? DEFAULT_MARGIN) || 0),
        bottom: Math.max(0, Number.parseFloat(margins?.bottom ?? DEFAULT_MARGIN) || 0)
    }

    return `${safe.left}:${safe.right}:${safe.top}:${safe.bottom}`
}

export const getUsableWidth = (pageWidth, margins = parsePageMargins()) => pageWidth - margins.left - margins.right

export const parseProps = (propsString) => {
    if (!propsString) return { font: 'Helvetica', size: 12, style: '000', align: 'left', borders: [0, 0, 0, 0] }
    const parts = propsString.split(':')
    return {
        font: parts[0] || 'Helvetica',
        size: parseInt(parts[1]) || 12,
        style: parts[2] || '000',
        align: parts[3] || 'left',
        borders: [
            parseInt(parts[4]) || 0,
            parseInt(parts[5]) || 0,
            parseInt(parts[6]) || 0,
            parseInt(parts[7]) || 0
        ]
    }
}

export const formatProps = (props) => {
    return `${props.font}:${props.size}:${props.style}:${props.align}:${props.borders.join(':')}`
}

export const parsePageBorder = (borderString) => {
    if (!borderString) return [0, 0, 0, 0]
    const parts = borderString.split(':')
    return [
        parseInt(parts[0]) || 0,
        parseInt(parts[1]) || 0,
        parseInt(parts[2]) || 0,
        parseInt(parts[3]) || 0
    ]
}

export const formatPageBorder = (borders) => {
    return borders.join(':')
}

// Helper function to get CSS font family from font name
export const getFontFamily = (fontName) => {
    if (!fontName) return 'Helvetica, Arial, sans-serif'

    const fontMap = {
        'Helvetica': 'Helvetica, Arial, sans-serif',
        'Helvetica-Bold': 'Helvetica, Arial, sans-serif',
        'Helvetica-Oblique': 'Helvetica, Arial, sans-serif',
        'Helvetica-BoldOblique': 'Helvetica, Arial, sans-serif',
        'Times-Roman': 'Times New Roman, Times, serif',
        'Times-Bold': 'Times New Roman, Times, serif',
        'Times-Italic': 'Times New Roman, Times, serif',
        'Times-BoldItalic': 'Times New Roman, Times, serif',
        'Courier': 'Courier New, Courier, monospace',
        'Courier-Bold': 'Courier New, Courier, monospace',
        'Courier-Oblique': 'Courier New, Courier, monospace',
        'Courier-BoldOblique': 'Courier New, Courier, monospace',
        'Symbol': 'Symbol, serif',
        'ZapfDingbats': 'ZapfDingbats, Wingdings, serif',
    }

    return fontMap[fontName] || `"${fontName}", sans-serif`
}

// Helper function to determine if font name implies bold/italic
export const getFontStyleFromName = (fontName) => {
    if (!fontName) return { isBold: false, isItalic: false }
    const lower = fontName.toLowerCase()
    return {
        isBold: lower.includes('bold'),
        isItalic: lower.includes('oblique') || lower.includes('italic')
    }
}

// Helper function to convert props to CSS style object
export const getStyleFromProps = (propsString) => {
    const parsed = parseProps(propsString)
    // Assuming fonts are handled appropriately, otherwise we might need access to custom fonts list
    // For basic utility usage, we rely on standard fonts or getFontFamily handling it.
    const fontStyles = getFontStyleFromName(parsed.font)

    const style = {
        fontSize: `${parsed.size}px`,
        textAlign: parsed.align,
        fontFamily: getFontFamily(parsed.font),
        borderLeftWidth: `${parsed.borders[0]}px`,
        borderRightWidth: `${parsed.borders[1]}px`,
        borderTopWidth: `${parsed.borders[2]}px`,
        borderBottomWidth: `${parsed.borders[3]}px`,
        borderStyle: 'solid',
        borderColor: '#333',
        fontWeight: 'normal',
        fontStyle: 'normal',
        textDecoration: 'none'
    }

    if (parsed.style[0] === '1' || fontStyles.isBold) {
        style.fontWeight = 'bold'
    }

    if (parsed.style[1] === '1' || fontStyles.isItalic) {
        style.fontStyle = 'italic'
    }

    if (parsed.style[2] === '1') {
        style.textDecoration = 'underline'
    }

    return style
}

// Helper to get image source with correct MIME type
export const getImageSrc = (imagedata, imagename) => {
    if (!imagedata) return ''
    if (imagedata.startsWith('data:')) return imagedata

    // Check for SVG extension or content signature
    const isSvg = (imagename && imagename.toLowerCase().endsWith('.svg')) ||
        imagedata.trim().startsWith('PHN2Zy') || // <svg (base64)
        imagedata.trim().startsWith('PD94bW')    // <?xm (base64)

    const mime = isSvg ? 'image/svg+xml' : 'image/png'
    return `data:${mime};base64,${imagedata}`
}
