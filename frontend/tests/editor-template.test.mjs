// Layer-1 tests: template build/parse/normalize/validate in
// src/components/editor/documentModel.js. Run with `npm test`.
import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

import {
  buildTemplate,
  buildTemplateJson,
  createFooter,
  createSpacer,
  createTable,
  createTitle,
  DEFAULT_CONFIG,
  normalizeConfig,
  parseComponentId,
  parseTemplateData,
  parseTemplateJson,
  validateTemplate,
} from '../src/components/editor/documentModel.js'

describe('editor template build', () => {
  it('emits elements and drops empty title, footer, and bookmarks', () => {
    const out = buildTemplate({
      config: { ...DEFAULT_CONFIG },
      title: createTitle(),
      components: [createTable(1, 1), createSpacer(9)],
      footer: null,
      bookmarks: [],
    })
    assert.equal(out.elements.length, 2)
    assert.deepEqual(out.elements[0], {
      type: 'table',
      table: { type: undefined, maxcolumns: 1, rows: out.elements[0].table.rows },
    })
    assert.ok('title' in out)
    assert.ok(!('footer' in out))
    assert.ok(!('bookmarks' in out))
    const withMarks = buildTemplate({ config: {}, title: null, components: [], footer: null, bookmarks: [{ title: 'A', page: 1 }] })
    assert.deepEqual(withMarks.bookmarks, [{ title: 'A', page: 1 }])
  })

  it('round-trips state through JSON without losing components', () => {
    const state = {
      config: { ...DEFAULT_CONFIG, pdfTitle: 'T' },
      title: createTitle(),
      components: [createTable(2, 2), createSpacer(12)],
      footer: createFooter(),
      bookmarks: [{ title: 'A', page: 2, children: [{ title: 'B', page: 3 }] }],
    }
    const json = buildTemplateJson(state)
    const back = parseTemplateData(JSON.parse(json), DEFAULT_CONFIG)
    assert.equal(back.components.length, 2)
    assert.equal(back.components[0].type, 'table')
    assert.equal(back.components[1].type, 'spacer')
    assert.equal(back.bookmarks[0].children[0].title, 'B')
    assert.equal(back.config.pdfTitle, 'T')
  })

  it('handles empty component lists instead of throwing', () => {
    const out = buildTemplate({ config: {}, title: null, components: null, footer: null })
    assert.deepEqual(out.elements, [])
  })
})

describe('editor template parse', () => {
  it('rejects trailing garbage in component ids', () => {
    assert.ok(Number.isNaN(parseComponentId('table-2-extra').index))
    assert.ok(Number.isNaN(parseComponentId('table-2x').index))
  })

  it('keeps canonical elements when legacy table keys are also present', () => {
    const parsed = parseTemplateData(
      {
        elements: [{ type: 'spacer', height: 5 }],
        table: [{ maxcolumns: 1, rows: [] }],
      },
      DEFAULT_CONFIG,
    )
    assert.ok(parsed.components.some((c) => c.type === 'spacer' && c.height === 5))
  })

  it('resolves indexed image refs from the image array', () => {
    const parsed = parseTemplateData(
      {
        elements: [{ type: 'image', index: 0 }],
        image: [{ width: 200, height: 150, imagename: 'a' }],
      },
      DEFAULT_CONFIG,
    )
    assert.equal(parsed.components.length, 1)
    assert.equal(parsed.components[0].width, 200)
  })

  it('parses JSON text into template state', () => {
    const parsed = parseTemplateJson(
      JSON.stringify({ config: { page: 'LETTER' }, elements: [{ type: 'spacer', height: 4 }], footer: { text: 'F' } }),
      DEFAULT_CONFIG,
    )
    assert.equal(parsed.config.page, 'LETTER')
    assert.equal(parsed.components[0].height, 4)
    assert.equal(parsed.footer.text, 'F')
  })
})

describe('editor config normalization', () => {
  it('keeps previous values when incoming keys are undefined', () => {
    const out = normalizeConfig({ page: undefined }, DEFAULT_CONFIG)
    assert.equal(out.page, 'A4')
  })

  it('canonicalizes the legacy embedFonts alias without leaking it', () => {
    const out = normalizeConfig({ embedFonts: false }, DEFAULT_CONFIG)
    assert.equal(out.embedStandardFonts, false)
    assert.ok(!('embedFonts' in out))
  })

  it('prefers embedStandardFonts and guards the known booleans', () => {
    const out = normalizeConfig({ embedStandardFonts: true, embedFonts: false, pdfaCompliant: false }, DEFAULT_CONFIG)
    assert.equal(out.embedStandardFonts, true)
    const kept = normalizeConfig({}, { ...DEFAULT_CONFIG, arlingtonCompatible: true })
    assert.equal(kept.arlingtonCompatible, true)
  })
})

describe('editor template validation', () => {
  it('accepts a well-formed template', () => {
    const { errors, warnings } = validateTemplate(
      JSON.parse(buildTemplateJson({
        config: { ...DEFAULT_CONFIG },
        title: createTitle(),
        components: [createTable(1, 1)],
        footer: createFooter(),
      })),
    )
    assert.deepEqual(errors, [])
    assert.deepEqual(warnings, [])
  })

  it('flags malformed margins, borders, and footer props', () => {
    const badMargin = validateTemplate({ config: { pageMargin: 'nope' }, elements: [] })
    assert.ok(badMargin.errors.some((e) => e.includes('pageMargin')))
    const badBorder = validateTemplate({ config: { pageBorder: '1:1' }, elements: [] })
    assert.ok(badBorder.errors.some((e) => e.includes('pageBorder')))
    const badFooter = validateTemplate({ config: {}, elements: [], footer: { props: 'garbage', text: 'f' } })
    assert.ok(badFooter.errors.some((e) => e.includes('footer.props')))
    const badFooterText = validateTemplate({ config: {}, elements: [], footer: { props: 'Helvetica:10:000:center:1:0:0:0', text: 7 } })
    assert.ok(badFooterText.errors.some((e) => e.includes('footer.text')))
  })

  it('warns on unknown page sizes', () => {
    const { warnings } = validateTemplate({ config: { page: 'B6' }, elements: [] })
    assert.ok(warnings.some((w) => w.includes('B6')))
  })

  it('reports border errors even when alignment is also unknown', () => {
    const { errors } = validateTemplate({
      config: {},
      elements: [],
      title: { props: 'Helvetica:12:000:justify:1:1', text: 'T' },
    })
    assert.ok(errors.some((e) => e.includes('title.props')))
  })
})
