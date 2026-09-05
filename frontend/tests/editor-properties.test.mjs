// Layer-1 tests: every editor property surface that is reachable from
// plain modules (documentModel.js + utils.js). Panel-only helpers living
// in .jsx files (bookmark dest sync, preset button maps) are covered
// indirectly through the strings they read and write. Run with `npm test`.
import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

import {
  BORDER_PATTERN,
  buildTemplateJson,
  createTable,
  DEFAULT_CELL_PROPS,
  insertComponent,
  KNOWN_PAGES,
  MARGIN_PATTERN,
  normalizeConfig,
  DEFAULT_CONFIG,
  PROPS_PATTERN,
  parseTemplateData,
  reorderComponentsByIndex,
  updateComponentById,
  formatPageMargins,
  formatProps,
  parsePageMargins,
  parseProps,
} from '../src/components/editor/documentModel.js'

describe('editor text props string (PropsEditor contract)', () => {
  it('round-trips font, size, style, align, and borders', () => {
    for (const sample of [
      'Helvetica:12:000:left:1:1:1:1',
      'Helvetica:18:100:center:1:1:1:1',
      'Times-Bold:10:010:right:0:0:0:1',
      'Courier:14:101:left:2:2:2:2',
    ]) {
      assert.equal(formatProps(parseProps(sample)), sample)
    }
  })

  it('unpacks each sub-field the panel edits', () => {
    const parsed = parseProps('Helvetica:18:100:center:1:1:1:1')
    assert.equal(parsed.font, 'Helvetica')
    assert.equal(parsed.size, 18)
    assert.equal(parsed.style, '100')
    assert.equal(parsed.align, 'center')
    assert.deepEqual(parsed.borders, [1, 1, 1, 1])
  })

  it('flips style bits and swaps alignment through formatProps', () => {
    const parsed = parseProps(DEFAULT_CELL_PROPS)
    const chars = parsed.style.split('')
    chars[0] = chars[0] === '1' ? '0' : '1'
    assert.equal(formatProps({ ...parsed, style: chars.join('') }), 'Helvetica:12:100:left:1:1:1:1')
    assert.equal(formatProps({ ...parsed, align: 'right' }), 'Helvetica:12:000:right:1:1:1:1')
  })

  it('accepts the border presets the panel offers (L:R:T:B order)', () => {
    assert.ok(PROPS_PATTERN.test('Helvetica:12:000:left:0:0:0:0'))
    assert.ok(PROPS_PATTERN.test('Helvetica:12:000:left:1:1:1:1'))
    assert.ok(PROPS_PATTERN.test('Helvetica:12:000:left:0:0:0:1'))
    assert.deepEqual(parseProps('Helvetica:12:000:left:0:0:0:1').borders, [0, 0, 0, 1])
  })

  it('falls back to Helvetica 12 plain left on empty input', () => {
    assert.deepEqual(parseProps(''), { font: 'Helvetica', size: 12, style: '000', align: 'left', borders: [0, 0, 0, 0] })
  })
})

describe('editor margins and page borders', () => {
  it('parses per-side margins with per-side fallback to 72', () => {
    assert.deepEqual(parsePageMargins('72:72:72:72'), { left: 72, right: 72, top: 72, bottom: 72 })
    assert.deepEqual(parsePageMargins('0:36:72:10.5'), { left: 0, right: 36, top: 72, bottom: 10.5 })
    const bad = parsePageMargins('-5:nope::')
    assert.deepEqual(bad, { left: 72, right: 72, top: 72, bottom: 72 })
    assert.deepEqual(parsePageMargins(null), { left: 72, right: 72, top: 72, bottom: 72 })
  })

  it('formats margins clamping negatives to zero', () => {
    assert.equal(formatPageMargins({ left: 0, right: 36, top: 72, bottom: 72 }), '0:36:72:72')
    assert.equal(formatPageMargins({ left: -4, right: 1, top: 2, bottom: 3 }), '0:1:2:3')
  })

  it('round-trips margin strings', () => {
    assert.equal(formatPageMargins(parsePageMargins('12:24:36:48')), '12:24:36:48')
  })

  it('validates the margin and border string shapes', () => {
    assert.ok(MARGIN_PATTERN.test('72:72:72:72'))
    assert.ok(!MARGIN_PATTERN.test('72:72:72'))
    assert.ok(!MARGIN_PATTERN.test('-1:0:0:0'))
    assert.ok(BORDER_PATTERN.test('0:0:1:0'))
    assert.ok(!BORDER_PATTERN.test('1:1:1'))
  })

  it('lists the supported page sizes', () => {
    assert.deepEqual(KNOWN_PAGES, ['A4', 'LETTER', 'LEGAL', 'A3', 'A5'])
  })
})

describe('editor document config properties', () => {
  it('starts from the documented defaults', () => {
    assert.deepEqual(DEFAULT_CONFIG, {
      pageBorder: '1:1:1:1',
      pageMargin: '72:72:72:72',
      page: 'A4',
      pageAlignment: 1,
      watermark: '',
      pdfTitle: '',
      pdfaCompliant: true,
    })
  })

  it('carries security and signature objects through normalization', () => {
    const out = normalizeConfig(
      { security: { enabled: true, ownerPassword: 'o', allowPrinting: false }, signature: { enabled: true, page: 2 } },
      DEFAULT_CONFIG,
    )
    assert.equal(out.security.enabled, true)
    assert.equal(out.security.allowPrinting, false)
    assert.equal(out.signature.page, 2)
    assert.equal(out.page, 'A4')
  })

  it('round-trips config through parse', () => {
    const parsed = parseTemplateData(
      { config: { page: 'LEGAL', pageAlignment: 2, watermark: 'draft', pdfTitle: 'R', pdfaCompliant: false }, elements: [] },
      DEFAULT_CONFIG,
    )
    assert.equal(parsed.config.page, 'LEGAL')
    assert.equal(parsed.config.pageAlignment, 2)
    assert.equal(parsed.config.watermark, 'draft')
    assert.equal(parsed.config.pdfaCompliant, false)
  })
})

describe('editor end-to-end property flow', () => {
  it('drop, style, reorder, and rename surface in the emitted JSON', () => {
    let components = []
    components = insertComponent(components, { ...createTable(1, 2), tag: 't1' }, null)
    components = insertComponent(components, { type: 'spacer', height: 20 }, null)
    components = updateComponentById(components, 'table-0', { bgcolor: '#E3F2FD' })
    const cell = components[0].rows[0].row[0]
    const styled = formatProps({ ...parseProps(cell.props), style: '100', align: 'center' })
    const rows = [{ row: [{ ...cell, props: styled, text: 'Hello' }] }, ...components[0].rows.slice(1)]
    components = updateComponentById(components, 'table-0', { rows })
    components = reorderComponentsByIndex(components, 1, 0)
    const json = buildTemplateJson({ config: { ...DEFAULT_CONFIG }, title: null, components, footer: null })
    const parsed = JSON.parse(json)
    assert.equal(parsed.elements[0].type, 'spacer')
    const tableEntry = parsed.elements[1].table
    assert.equal(tableEntry.bgcolor, '#E3F2FD')
    assert.equal(tableEntry.rows[0].row[0].props, 'Helvetica:12:100:center:1:1:1:1')
    assert.equal(tableEntry.rows[0].row[0].text, 'Hello')
  })
})
