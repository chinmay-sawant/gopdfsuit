// Pure-function tests for the Editor document reducers in
// src/components/editor/documentModel.js (Phase 3.6). Run with `npm test`.
import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

import {
  applyCellDropToRows,
  cellDropData,
  createTable,
  deleteComponentById,
  findElementById,
  insertComponent,
  moveComponentByIndex,
  parseComponentId,
  pasteComponentAt,
  reorderComponentsByIndex,
  updateComponentById,
} from '../src/components/editor/documentModel.js'

const table = (tag) => ({ ...createTable(1, 1), tag })
const spacer = (tag) => ({ type: 'spacer', height: 10, tag })

describe('documentModel reducers', () => {
  it('parses component ids', () => {
    assert.deepEqual(parseComponentId('title'), { kind: 'title', index: -1 })
    assert.deepEqual(parseComponentId('table-2'), { kind: 'table', index: 2 })
    assert.ok(Number.isNaN(parseComponentId('bogus').index))
  })

  it('finds elements across title/components/footer', () => {
    const state = { title: { text: 'T' }, components: [spacer('a')], footer: { text: 'F' } }
    assert.equal(findElementById(state, 'title').type, 'title')
    assert.equal(findElementById(state, 'spacer-0').tag, 'a')
    assert.equal(findElementById(state, 'footer').type, 'footer')
    assert.equal(findElementById(state, 'table-9'), null)
  })

  it('inserts before the target id and appends otherwise', () => {
    const list = [table('a'), table('c')]
    const inserted = insertComponent(list, table('b'), 'table-1')
    assert.deepEqual(inserted.map((c) => c.tag), ['a', 'b', 'c'])
    assert.deepEqual(insertComponent(list, table('z'), 'table-99').map((c) => c.tag), ['a', 'c', 'z'])
    assert.deepEqual(insertComponent(list, table('z')).map((c) => c.tag), ['a', 'c', 'z'])
    assert.equal(list.length, 2)
  })

  it('deletes and updates by id without mutating', () => {
    const list = [table('a'), spacer('b')]
    assert.deepEqual(deleteComponentById(list, 'spacer-1'), [list[0]])
    assert.equal(deleteComponentById(list, 'table-9'), list)
    const updated = updateComponentById(list, 'table-0', { tag: 'a2' })
    assert.equal(updated[0].tag, 'a2')
    assert.equal(list[0].tag, 'a')
  })

  it('moves components up and down', () => {
    const list = [table('a'), table('b'), table('c')]
    assert.deepEqual(moveComponentByIndex(list, 1, 'up').map((c) => c.tag), ['b', 'a', 'c'])
    assert.deepEqual(moveComponentByIndex(list, 1, 'down').map((c) => c.tag), ['a', 'c', 'b'])
    assert.deepEqual(moveComponentByIndex(list, 0, 'up').map((c) => c.tag), ['a', 'b', 'c'])
  })

  it('reorders by index and rejects bad indices', () => {
    const list = [table('a'), table('b'), table('c')]
    assert.deepEqual(reorderComponentsByIndex(list, 0, 2).map((c) => c.tag), ['b', 'c', 'a'])
    assert.equal(reorderComponentsByIndex(list, 1, 1), list)
    assert.equal(reorderComponentsByIndex(list, 0, 9), list)
  })

  it('pastes after the target id', () => {
    const list = [table('a'), table('b')]
    assert.deepEqual(pasteComponentAt(list, table('x'), 'table-0').map((c) => c.tag), ['a', 'x', 'b'])
    assert.deepEqual(pasteComponentAt(list, table('x'), 'title').map((c) => c.tag), ['a', 'b', 'x'])
    assert.deepEqual(pasteComponentAt(list, table('x')).map((c) => c.tag), ['a', 'b', 'x'])
  })

  it('builds cell-drop payloads per field type', () => {
    assert.equal(cellDropData('checkbox', 7).form_field.name, 'checkbox_7')
    assert.equal(cellDropData('text_input', 7).form_field.type, 'text')
    assert.equal(cellDropData('hyperlink', 7).text, 'Link Text')
    assert.equal(cellDropData('nope', 7), null)
  })

  it('applies a cell drop to a row copy', () => {
    const rows = [{ row: [{ props: 'p', text: 'keep' }] }]
    const next = applyCellDropToRows(rows, 0, 0, 'checkbox', 7)
    assert.equal(next[0].row[0].form_field.name, 'checkbox_7')
    assert.equal(rows[0].row[0].text, 'keep')
    assert.equal(applyCellDropToRows(rows, 5, 0, 'checkbox', 7), rows)
  })
})
