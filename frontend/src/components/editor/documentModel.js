import {
  DEFAULT_MARGIN,
  formatPageMargins,
  formatProps,
  parsePageMargins,
  parseProps,
} from './utils.js'

export { DEFAULT_MARGIN, formatPageMargins, formatProps, parsePageMargins, parseProps }

export const DEFAULT_CONFIG = {
  pageBorder: '1:1:1:1',
  pageMargin: '72:72:72:72',
  page: 'A4',
  pageAlignment: 1,
  watermark: '',
  pdfTitle: '',
  pdfaCompliant: true,
}

export const DEFAULT_CELL_PROPS = 'Helvetica:12:000:left:1:1:1:1'
export const DEFAULT_FLAT_CELL_PROPS = 'Helvetica:12:000:left:0:0:0:0'

export const createTitle = (overrides = {}) => ({
  props: 'Helvetica:12:000:left:1:1:1:1',
  text: 'Document Title',
  textprops: 'Helvetica:18:100:center:1:1:1:1',
  table: {
    maxcolumns: 3,
    columnwidths: [1, 2, 1],
    rows: [{
      row: [
        { props: 'Helvetica:12:000:left:1:1:1:1', text: '', image: null },
        { props: 'Helvetica:18:100:center:1:1:1:1', text: 'Document Title' },
        { props: 'Helvetica:12:000:right:1:1:1:1', text: '' },
      ],
    }],
  },
  ...overrides,
})

export const createFooter = (overrides = {}) => ({
  props: 'Helvetica:10:000:center:1:0:0:0',
  text: 'Page footer text',
  ...overrides,
})

export const createTableCell = (overrides = {}) => ({
  props: DEFAULT_CELL_PROPS,
  text: '',
  ...overrides,
})

export const createTable = (rows = 3, cols = 3, cellProps = DEFAULT_CELL_PROPS) => ({
  type: 'table',
  maxcolumns: cols,
  rows: Array.from({ length: rows }, () => ({
    row: Array.from({ length: cols }, () => createTableCell({ props: cellProps })),
  })),
})

export const createSpacer = (height = 20) => ({ type: 'spacer', height })

export const createImage = (overrides = {}) => ({
  type: 'image',
  width: 200,
  height: 150,
  imagedata: null,
  imagename: '',
  ...overrides,
})

export const createComponent = (type) => {
  if (type === 'table') return createTable()
  if (type === 'image') return createImage()
  return createSpacer()
}

export const wrapComponent = (component) => {
  if (!component) return component
  if (component.type === 'table') return { type: 'table', table: { ...component, type: undefined } }
  if (component.type === 'spacer') return { type: 'spacer', spacer: { ...component, type: undefined } }
  if (component.type === 'image') return { type: 'image', image: { ...component, type: undefined } }
  return component
}

const unwrapComponent = (entry) => {
  if (!entry || typeof entry !== 'object') return entry
  if (entry.table) return { ...entry.table, type: 'table' }
  if (entry.spacer) return { ...entry.spacer, type: 'spacer' }
  if (entry.image) return { ...entry.image, type: 'image' }
  if (!entry.type) {
    if (entry.maxcolumns && entry.rows) return { ...entry, type: 'table' }
    if (entry.height !== undefined && entry.width === undefined) return { ...entry, type: 'spacer' }
    if (entry.imagedata || entry.imagename) return { ...entry, type: 'image' }
  }
  return entry
}

export const buildTemplate = ({ config, title, components = [], footer, bookmarks } = {}) => {
  const template = {
    config: { ...config },
    title,
    elements: components.map(wrapComponent),
    footer,
    bookmarks,
  }
  if (!title) delete template.title
  if (!footer) delete template.footer
  if (!bookmarks || bookmarks.length === 0) delete template.bookmarks
  return template
}

export const buildTemplateJson = (state) => JSON.stringify(buildTemplate(state), null, 2)

export const normalizeConfig = (incoming = {}, prev = DEFAULT_CONFIG) => {
  const embedValue = incoming.embedStandardFonts !== undefined
    ? incoming.embedStandardFonts
    : (incoming.embedFonts !== undefined ? incoming.embedFonts : undefined)
  return {
    ...prev,
    ...incoming,
    embedStandardFonts: embedValue !== undefined ? embedValue : prev.embedStandardFonts,
    arlingtonCompatible: incoming.arlingtonCompatible !== undefined ? incoming.arlingtonCompatible : prev.arlingtonCompatible,
    pdfaCompliant: incoming.pdfaCompliant !== undefined ? incoming.pdfaCompliant : prev.pdfaCompliant,
  }
}

export const parseTemplateData = (data = {}, prevConfig = DEFAULT_CONFIG) => {
  const { config: newConfig, title: newTitle, elements, table, spacer, content, footer: newFooter, bookmarks: newBookmarks } = data

  let rawComponents = elements || content || []
  if (table && Array.isArray(table)) {
    rawComponents = table.map((entry) => ({ ...entry, type: 'table' }))
  }
  if (spacer && Array.isArray(spacer)) {
    rawComponents = [...rawComponents, ...spacer.map((entry) => ({ ...entry, type: 'spacer' }))]
  }
  if (data.elements && Array.isArray(data.elements) && data.elements[0]?.index !== undefined) {
    const ordered = []
    for (const ref of data.elements) {
      if (ref.type === 'table' && table && table[ref.index]) ordered.push({ ...table[ref.index], type: 'table' })
      else if (ref.type === 'spacer' && spacer && spacer[ref.index]) ordered.push({ ...spacer[ref.index], type: 'spacer' })
    }
    if (ordered.length > 0) rawComponents = ordered
  }

  const components = (Array.isArray(rawComponents) ? rawComponents : []).map(unwrapComponent)
  return {
    config: normalizeConfig(newConfig || {}, prevConfig),
    title: newTitle || null,
    components,
    footer: newFooter || null,
    bookmarks: newBookmarks || null,
  }
}

export const parseTemplateJson = (jsonText, prevConfig = DEFAULT_CONFIG) => parseTemplateData(JSON.parse(jsonText), prevConfig)

// --- Validation against frontend/template.schema.json rules ---

export const PROPS_PATTERN = /^[^:]+:\d+:[01]{3}:(left|center|right):\d+:\d+:\d+:\d+$/
export const MARGIN_PATTERN = /^\d+(\.\d+)?:\d+(\.\d+)?:\d+(\.\d+)?:\d+(\.\d+)?$/
export const BORDER_PATTERN = /^\d+:\d+:\d+:\d+$/
export const KNOWN_PAGES = ['A4', 'LETTER', 'LEGAL', 'A3', 'A5']

const isKnownPage = (page) => !page || KNOWN_PAGES.includes(String(page).toUpperCase())

const checkProps = (value, path, errors, warnings) => {
  if (typeof value !== 'string' || !value) {
    errors.push(`${path}: missing props string`)
    return
  }
  if (PROPS_PATTERN.test(value)) return
  const parts = value.split(':')
  if (parts.length >= 4 && !['left', 'center', 'right'].includes(parts[3])) {
    warnings.push(`${path}: unknown alignment "${parts[3]}" (engine falls back to left)`)
    return
  }
  errors.push(`${path}: malformed props "${value}" (want font:size:style:align:borders)`)
}

const checkCell = (cell, path, errors, warnings) => {
  if (!cell || typeof cell !== 'object') {
    errors.push(`${path}: cell must be an object`)
    return
  }
  checkProps(cell.props, `${path}.props`, errors, warnings)
}

const checkTable = (table, path, errors, warnings) => {
  if (!table || typeof table !== 'object') {
    errors.push(`${path}: table must be an object`)
    return
  }
  if (!Number.isInteger(table.maxcolumns) || table.maxcolumns < 1) {
    errors.push(`${path}.maxcolumns: must be a positive integer`)
  }
  if (!Array.isArray(table.rows)) {
    errors.push(`${path}.rows: must be an array`)
    return
  }
  table.rows.forEach((row, rowIdx) => {
    if (!row || !Array.isArray(row.row)) {
      errors.push(`${path}.rows[${rowIdx}]: must have a "row" cell array`)
      return
    }
    if (Number.isInteger(table.maxcolumns) && row.row.length !== table.maxcolumns) {
      warnings.push(`${path}.rows[${rowIdx}]: ${row.row.length} cells but maxcolumns=${table.maxcolumns}`)
    }
    row.row.forEach((cell, colIdx) => checkCell(cell, `${path}.rows[${rowIdx}].row[${colIdx}]`, errors, warnings))
  })
}

export const validateTemplate = (template = {}) => {
  const errors = []
  const warnings = []
  if (!template || typeof template !== 'object') return { errors: ['template must be an object'], warnings }

  const config = template.config || {}
  if (config.pageMargin !== undefined && !MARGIN_PATTERN.test(String(config.pageMargin))) {
    errors.push(`config.pageMargin: malformed "${config.pageMargin}" (want left:right:top:bottom)`)
  }
  if (config.pageBorder !== undefined && !BORDER_PATTERN.test(String(config.pageBorder))) {
    errors.push(`config.pageBorder: malformed "${config.pageBorder}"`)
  }
  if (!isKnownPage(config.page)) warnings.push(`config.page: unknown size "${config.page}" (engine defaults to A4)`)

  if (template.title) {
    checkProps(template.title.props, 'title.props', errors, warnings)
    if (template.title.table) checkTable(template.title.table, 'title.table', errors, warnings)
  }

  const elements = template.elements || []
  if (!Array.isArray(elements)) {
    errors.push('elements: must be an array')
  } else {
    elements.forEach((entry, idx) => {
      const path = `elements[${idx}]`
      if (!entry || typeof entry !== 'object') {
        errors.push(`${path}: entry must be an object`)
        return
      }
      if (entry.table) checkTable({ ...entry.table, type: undefined }, `${path}.table`, errors, warnings)
      else if (entry.spacer) {
        if (typeof entry.spacer.height !== 'number') errors.push(`${path}.spacer.height: must be a number`)
      } else if (entry.image) {
        if (typeof entry.image.width !== 'number' || typeof entry.image.height !== 'number') {
          errors.push(`${path}.image: width/height must be numbers`)
        }
      } else {
        errors.push(`${path}: entry must wrap table, spacer, or image`)
      }
    })
  }

  if (template.footer && typeof template.footer.text !== 'string') {
    errors.push('footer.text: must be a string')
  }
  return { errors, warnings }
}

export default {
  buildTemplate,
  buildTemplateJson,
  parseTemplateData,
  parseTemplateJson,
  validateTemplate,
}
