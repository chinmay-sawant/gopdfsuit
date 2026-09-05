// Layer-1 tests: palette-to-cell drops in
// src/components/editor/documentModel.js. Run with `npm test`.
import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

import {
  applyCellDropToRows,
  cellDropData,
} from '../src/components/editor/documentModel.js'

const rowsOf = (cell) => [{ row: [cell] }]
const styledCell = () => ({
  props: 'Helvetica:12:100:center:1:1:1:1',
  text: 'hi',
  width: 120,
  bgcolor: '#E3F2FD',
  link: 'https://example.com',
  dest: 'sec-1',
})

describe('editor cell drop payloads', () => {
  it('builds a checkbox field with a stable name per stamp', () => {
    const data = cellDropData('checkbox', 7)
    assert.deepEqual(data.form_field, { name: 'checkbox_7', checked: false, type: 'checkbox' })
    assert.equal(data.props, 'Helvetica:12:000:left:0:0:0:0')
  })

  it('builds simple checkbox and radio flags', () => {
    assert.equal(cellDropData('checkbox_simple', 1).chequebox, false)
    assert.equal(cellDropData('radio_simple', 1).radio, false)
  })

  it('builds text, radio, image, and hyperlink payloads', () => {
    assert.deepEqual(cellDropData('text_input', 3).form_field, { name: 'field_3', value: '', type: 'text' })
    assert.deepEqual(cellDropData('radio', 4).form_field, { name: 'radio_4', checked: false, type: 'radio' })
    assert.deepEqual(cellDropData('image', 5).image, { imagename: '', imagedata: null, width: 100, height: 80 })
    const link = cellDropData('hyperlink', 6)
    assert.equal(link.text, 'Link Text')
    assert.equal(link.link, 'https://example.com')
  })

  it('returns null for unknown drop types', () => {
    assert.equal(cellDropData('title', 1), null)
    assert.equal(cellDropData('table', 1), null)
  })
})

describe('editor cell drop application', () => {
  it('applies the drop to a copy without mutating the input', () => {
    const rows = rowsOf({ props: 'Helvetica:12:000:left:1:1:1:1', text: 'a' })
    const out = applyCellDropToRows(rows, 0, 0, 'checkbox', 9)
    assert.equal(out[0].row[0].form_field.name, 'checkbox_9')
    assert.equal(rows[0].row[0].text, 'a')
    assert.ok(!('form_field' in rows[0].row[0]))
    assert.notEqual(out, rows)
    assert.notEqual(out[0].row, rows[0].row)
  })

  it('keeps the existing cell styling when a field is dropped', () => {
    const out = applyCellDropToRows(rowsOf(styledCell()), 0, 0, 'checkbox', 1)
    const cell = out[0].row[0]
    assert.equal(cell.form_field.type, 'checkbox')
    assert.equal(cell.width, 120)
    assert.equal(cell.bgcolor, '#E3F2FD')
    assert.equal(cell.link, 'https://example.com')
    assert.equal(cell.dest, 'sec-1')
  })

  it('leaves rows untouched for bad targets and unknown types', () => {
    const rows = rowsOf({ props: 'Helvetica:12:000:left:1:1:1:1', text: 'a' })
    assert.equal(applyCellDropToRows(rows, 5, 0, 'checkbox', 1), rows)
    assert.equal(applyCellDropToRows(rows, 0, 3, 'checkbox', 1), rows)
    assert.equal(applyCellDropToRows(rows, 0, 0, 'title', 1), rows)
  })

  it('leaves malformed rows untouched instead of spreading them', () => {
    const rows = [{ row: 'abc' }]
    assert.equal(applyCellDropToRows(rows, 0, 0, 'checkbox', 7), rows)
  })
})
