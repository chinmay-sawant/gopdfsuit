import { parseProps } from './utils.js'

const goString = (value) => JSON.stringify(value ?? '')
const pyString = (value) => JSON.stringify(value ?? '')

const goBool = (value) => (value ? 'true' : 'false')
const pyBool = (value) => (value ? 'True' : 'False')

const styleFlags = (style = '000') => ({
    bold: style[0] === '1',
    italic: style[1] === '1',
    underline: style[2] === '1',
})

const safeBorders = (borders) => (
    Array.isArray(borders) && borders.length === 4
        ? borders.map((b) => Number(b) || 0)
        : [0, 0, 0, 0]
)

const clampSpacerHeight = (height) => {
    const n = Math.round(Number(height))
    if (!Number.isFinite(n)) return 20
    return Math.min(200, Math.max(1, n))
}

const clampImageDim = (value, fallback) => {
    const n = Math.round(Number(value))
    if (!Number.isFinite(n) || n <= 0) return fallback
    return n
}

const hasBrackets = (text) => (
    typeof text === 'string' && (/\[.*\]/.test(text) || /\(.*\)/.test(text))
)

const goFontChain = (cell = {}) => {
    const p = parseProps(cell.props)
    const f = styleFlags(p.style)
    const borders = safeBorders(p.borders)
    const parts = [`gopdflib.Font(${goString(p.font)})`, `Size(${p.size})`]
    if (f.bold) parts.push('Bold()')
    if (f.italic) parts.push('Italic()')
    if (f.underline) parts.push('Underline()')
    if (p.align === 'center') parts.push('Center()')
    else if (p.align === 'right') parts.push('Right()')
    if (borders.every((b) => b === 1)) parts.push('Bordered()')
    else if (!borders.every((b) => b === 0)) parts.push(`Borders(${borders.join(', ')})`)
    return parts.join('.')
}

const pyFontChain = (cell = {}) => {
    const p = parseProps(cell.props)
    const f = styleFlags(p.style)
    const borders = safeBorders(p.borders)
    const parts = [`Font(${pyString(p.font)})`, `size(${p.size})`]
    if (f.bold) parts.push('bold()')
    if (f.italic) parts.push('italic()')
    if (f.underline) parts.push('underline()')
    if (p.align === 'center') parts.push('center()')
    else if (p.align === 'right') parts.push('right()')
    if (borders.every((b) => b === 1)) parts.push('bordered()')
    else if (borders.every((b) => b === 0)) parts.push('borderless()')
    else parts.push(`borders(${borders.join(', ')})`)
    return parts.join('.')
}

const goCellExpr = (cell = {}) => `${goFontChain(cell)}.Cell(${goString(cell.text ?? '')})`
const pyCellExpr = (cell = {}) => `${pyFontChain(cell)}.cell(${pyString(cell.text ?? '')})`

const imageTag = (image = {}) => {
    const name = image.imagename || image.ImageName || ''
    const w = image.width ?? image.Width ?? ''
    const h = image.height ?? image.Height ?? ''
    return `${name} ${w}x${h}`.trim()
}

const goCellImageLines = (cell = {}, target) => {
    const lines = []
    const image = cell.image
    if (!image) return lines
    const name = image.imagename || ''
    const dataLen = typeof image.imagedata === 'string' ? image.imagedata.length : 0
    const goFloat = (n) => (Number.isInteger(n) ? `${n}.0` : String(n))
    const w = Number(image.width ?? cell.width)
    const h = Number(image.height ?? cell.height)
    if (Number.isFinite(w) && w > 0) lines.push(`imgW := ${goFloat(w)}
${target}.Width = &imgW`)
    if (Number.isFinite(h) && h > 0) lines.push(`imgH := ${goFloat(h)}
${target}.Height = &imgH`)
    lines.push(`${target}.Image = &gopdflib.Image{ImageName: ${goString(name)}, ImageData: "<base64 omitted${dataLen ? `, ${dataLen} chars` : ''}>"}`)
    return lines
}

export const cellToGoSnippet = (cell = {}) => {
    const lines = [`c := ${goCellExpr(cell)}`]
    if (cell.bgcolor) lines.push(`gopdflib.SetCellColor(&c, ${goString(cell.bgcolor)})`)
    if (cell.textcolor) lines.push(`gopdflib.SetCellTextColor(&c, ${goString(cell.textcolor)})`)
    if (cell.link) lines.push(`c.Link = ${goString(cell.link)}`)
    lines.push(...goCellImageLines(cell, 'c'))
    if (hasBrackets(cell.text)) lines.push('gopdflib.AddBracketText(&c, "[", "]")')
    return lines.join('\n')
}

export const cellToPythonSnippet = (cell = {}) => {
    const lines = [`c = ${pyCellExpr(cell)}`]
    if (cell.bgcolor) lines.push(`set_cell_color(c, ${pyString(cell.bgcolor)})`)
    if (cell.textcolor) lines.push(`set_cell_text_color(c, ${pyString(cell.textcolor)})`)
    if (cell.link) lines.push(`c["link"] = ${pyString(cell.link)}`)
    if (cell.image) {
        const name = cell.image.imagename || ''
        const w = clampImageDim(cell.image.width ?? cell.width, 100)
        const h = clampImageDim(cell.image.height ?? cell.height, 80)
        lines.push(`c["image"] = {"imagename": ${pyString(name)}, "width": ${w}, "height": ${h}}  # imagedata omitted`)
    }
    if (hasBrackets(cell.text)) lines.push('add_bracket_text(c, "[", "]")')
    return lines.join('\n')
}

const goTableLines = (table = {}, tbName, rowName) => {
    const rows = Array.isArray(table.rows) ? table.rows : []
    const maxcolumns = table.maxcolumns || rows[0]?.row?.length || 1
    const widths = Array.isArray(table.columnwidths) && table.columnwidths.length > 0
        ? table.columnwidths
        : Array.from({ length: maxcolumns }, () => 1)
    const lines = [`${tbName} := b.AddTable(${maxcolumns}, ${widths.join(', ')})`]
    rows.forEach((entry, rowIdx) => {
        const cells = Array.isArray(entry?.row) ? entry.row : []
        const rowVar = rows.length > 1 ? `${rowName}${rowIdx + 1}` : rowName
        lines.push(`${rowVar} := ${tbName}.AddRow(`)
        cells.forEach((cell, cellIdx) => {
            lines.push(`\t${goCellExpr(cell)}${cellIdx < cells.length - 1 ? ',' : ''}`)
        })
        lines.push(')')
        cells.forEach((cell, cellIdx) => {
            if (!cell) return
            if (cell.bgcolor) lines.push(`gopdflib.SetCellColor(&${rowVar}[${cellIdx}], ${goString(cell.bgcolor)})`)
            if (cell.textcolor) lines.push(`gopdflib.SetCellTextColor(&${rowVar}[${cellIdx}], ${goString(cell.textcolor)})`)
            if (cell.link) lines.push(`${rowVar}[${cellIdx}].Link = ${goString(cell.link)}`)
            if (cell.image) lines.push(`// ${rowVar}[${cellIdx}] image ${imageTag(cell.image)} (imagedata omitted)`)
            if (hasBrackets(cell.text)) lines.push(`gopdflib.AddBracketText(&${rowVar}[${cellIdx}], "[", "]")`)
        })
    })
    return lines
}

const pyTableLines = (table = {}, tbName, rowName) => {
    const rows = Array.isArray(table.rows) ? table.rows : []
    const maxcolumns = table.maxcolumns || rows[0]?.row?.length || 1
    const widths = Array.isArray(table.columnwidths) && table.columnwidths.length > 0
        ? table.columnwidths
        : Array.from({ length: maxcolumns }, () => 1)
    const lines = [`${tbName} = b.add_table(${maxcolumns}, ${widths.join(', ')})`]
    rows.forEach((entry, rowIdx) => {
        const cells = Array.isArray(entry?.row) ? entry.row : []
        const rowVar = rows.length > 1 ? `${rowName}${rowIdx + 1}` : rowName
        lines.push(`${rowVar} = ${tbName}.add_row(`)
        cells.forEach((cell, cellIdx) => {
            lines.push(`    ${pyCellExpr(cell)}${cellIdx < cells.length - 1 ? ',' : ''}`)
        })
        lines.push(')')
        cells.forEach((cell, cellIdx) => {
            if (!cell) return
            if (cell.bgcolor) lines.push(`set_cell_color(${rowVar}[${cellIdx}], ${pyString(cell.bgcolor)})`)
            if (cell.textcolor) lines.push(`set_cell_text_color(${rowVar}[${cellIdx}], ${pyString(cell.textcolor)})`)
            if (cell.link) lines.push(`${rowVar}[${cellIdx}]["link"] = ${pyString(cell.link)}`)
            if (cell.image) lines.push(`# ${rowVar}[${cellIdx}] image ${imageTag(cell.image)} (imagedata omitted)`)
            if (hasBrackets(cell.text)) lines.push(`add_bracket_text(${rowVar}[${cellIdx}], "[", "]")`)
        })
    })
    return lines
}

export const tableToGoSnippet = (table = {}) => goTableLines(table, 'tb', 'row').join('\n')
export const tableToPythonSnippet = (table = {}) => pyTableLines(table, 'tb', 'row').join('\n')

const unwrapElement = (entry) => {
    if (!entry || typeof entry !== 'object') return null
    if (entry.table) return { ...entry.table, type: 'table' }
    if (entry.spacer) return { ...entry.spacer, type: 'spacer' }
    if (entry.image) return { ...entry.image, type: 'image' }
    return entry
}

const normalizeElements = (template = {}) => {
    const raw = Array.isArray(template.elements)
        ? template.elements
        : (Array.isArray(template.components) ? template.components : [])
    return raw.map(unwrapElement).filter(Boolean)
}

const goTitleLines = (title) => {
    if (!title) return []
    const tp = parseProps(title.textprops || title.props)
    const f = styleFlags(tp.style)
    const lines = [`b.AddTitle(${goString(title.text ?? '')}, gopdflib.WithTitleFontOpts(gopdflib.TitleFontOptions{Name: ${goString(tp.font)}, Size: ${tp.size}, Bold: ${goBool(f.bold)}, Italic: ${goBool(f.italic)}, Underline: ${goBool(f.underline)}}))`]
    if (title.table) lines.push(...goTableLines(title.table, 'titleTb', 'titleRow'))
    return lines
}

const pyTitleLines = (title) => {
    if (!title) return []
    const tp = parseProps(title.textprops || title.props)
    const f = styleFlags(tp.style)
    const lines = [`b.add_title(${pyString(title.text ?? '')}, ${pyString(tp.font)}, ${tp.size}, ${pyBool(f.bold)}, ${pyBool(f.italic)}, ${pyBool(f.underline)})`]
    if (title.table) lines.push(...pyTableLines(title.table, 'title_tb', 'title_row'))
    return lines
}

export const templateToGoSnippet = (template = {}) => {
    const page = template.config?.page || 'A4'
    const lines = [
        '// import "github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"',
        `b := gopdflib.NewDocument(${goString(page)}, true)`,
    ]
    lines.push(...goTitleLines(template.title))
    let tableIdx = 0
    normalizeElements(template).forEach((entry) => {
        if (entry.type === 'table') {
            tableIdx += 1
            const tbName = tableIdx > 1 ? `tb${tableIdx}` : 'tb'
            lines.push(...goTableLines(entry, tbName, `row${tableIdx}`))
        } else if (entry.type === 'spacer') {
            lines.push(`b.AddSpacer(${clampSpacerHeight(entry.height)})`)
        } else if (entry.type === 'image') {
            const w = clampImageDim(entry.width, 200)
            const h = clampImageDim(entry.height, 150)
            const name = entry.imagename || ''
            const dataLen = typeof entry.imagedata === 'string' ? entry.imagedata.length : 0
            lines.push(`b.AddImage(${w}, ${h}, ${goString(name)}) // imagedata${dataLen ? ` ${dataLen} chars` : ''} omitted`)
        }
    })
    if (template.footer) {
        lines.push(`// footer: ${goString(template.footer.text ?? '')} (${template.footer.props ?? ''})`)
    }
    lines.push('pdfBytes, err := b.Generate()')
    return lines.join('\n')
}

export const templateToPythonSnippet = (template = {}) => {
    const page = template.config?.page || 'A4'
    const lines = [
        '# from pypdfsuit.builder import TemplateBuilder, Font, set_cell_color, set_cell_text_color, add_bracket_text',
        `b = TemplateBuilder(${pyString(page)}, True)`,
    ]
    lines.push(...pyTitleLines(template.title))
    let tableIdx = 0
    normalizeElements(template).forEach((entry) => {
        if (entry.type === 'table') {
            tableIdx += 1
            const tbName = tableIdx > 1 ? `tb${tableIdx}` : 'tb'
            lines.push(...pyTableLines(entry, tbName, `row${tableIdx}`))
        } else if (entry.type === 'spacer') {
            lines.push(`b.add_spacer(${clampSpacerHeight(entry.height)})`)
        } else if (entry.type === 'image') {
            const w = clampImageDim(entry.width, 200)
            const h = clampImageDim(entry.height, 150)
            const name = entry.imagename || ''
            const dataLen = typeof entry.imagedata === 'string' ? entry.imagedata.length : 0
            lines.push(`b.add_image(${w}, ${h}, ${pyString(name)})  # imagedata${dataLen ? ` ${dataLen} chars` : ''} omitted`)
        }
    })
    if (template.footer) {
        lines.push(`# footer: ${pyString(template.footer.text ?? '')} (${template.footer.props ?? ''})`)
    }
    lines.push('pdf_bytes = b.generate()')
    return lines.join('\n')
}

export default {
    cellToGoSnippet,
    cellToPythonSnippet,
    tableToGoSnippet,
    tableToPythonSnippet,
    templateToGoSnippet,
    templateToPythonSnippet,
}
