// Layer-1 tests: component creators and structural reducers in
// src/components/editor/documentModel.js. Run with `npm test`
// (node --test, no browser needed).
import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

import {
  createComponent,
  createFooter,
  createImage,
  createSpacer,
  createTable,
  createTableCell,
  createTitle,
  DEFAULT_CELL_PROPS,
  deleteComponentById,
  findElementById,
  insertComponent,
  moveComponentByIndex,
  pasteComponentAt,
  reorderComponentsByIndex,
  updateComponentById,
  wrapComponent,
} from '../src/components/editor/documentModel.js'

const table = (tag) => ({ ...createTable(1, 1), tag })
const spacer = (tag) => ({ type: 'spacer', height: 10, tag })

describe('editor component creators', () => {
  it('creates tables with the requested shape and default cell props', () => {
    const t = createTable(2, 3)
    assert.equal(t.type, 'table')
    assert.equal(t.maxcolumns, 3)
    assert.equal(t.rows.length, 2)
    assert.equal(t.rows[0].row.length, 3)
    assert.equal(t.rows[1].row[2].props, DEFAULT_CELL_PROPS)
  })

  it('creates title, footer, spacer, image, and cell defaults', () => {
    const title = createTitle()
    assert.equal(title.text, 'Document Title')
    assert.equal(title.table.maxcolumns, 3)
    assert.equal(createFooter().text, 'Page footer text')
    assert.deepEqual(createSpacer(), { type: 'spacer', height: 20 })
    assert.equal(createSpacer(44).height, 44)
    const image = createImage()
    assert.equal(image.type, 'image')
    assert.equal(image.width, 200)
    assert.equal(image.height, 150)
    assert.equal(image.imagedata, null)
    assert.equal(createTableCell().props, DEFAULT_CELL_PROPS)
  })

  it('creates a component per palette type with a spacer fallback', () => {
    assert.equal(createComponent('table').type, 'table')
    assert.equal(createComponent('image').type, 'image')
    assert.equal(createComponent('spacer').type, 'spacer')
    assert.equal(createComponent('chart').type, 'spacer')
  })

  it('wraps components for template JSON and passes unknowns through', () => {
    const wrapped = wrapComponent({ type: 'table', maxcolumns: 1, rows: [] })
    assert.deepEqual(wrapped, { type: 'table', table: { type: undefined, maxcolumns: 1, rows: [] } })
    assert.deepEqual(wrapComponent({ type: 'spacer', height: 5 }), { type: 'spacer', spacer: { type: undefined, height: 5 } })
    assert.equal(wrapComponent(null), null)
  })
})

describe('editor structural reducers', () => {
  it('inserts, deletes, and updates by id without mutating the input', () => {
    const list = [table('a'), table('b')]
    const inserted = insertComponent(list, table('x'), 'table-1')
    assert.deepEqual(inserted.map((c) => c.tag), ['a', 'x', 'b'])
    assert.equal(list.length, 2)
    const mixed = [table('a'), spacer('b')]
    assert.deepEqual(deleteComponentById(mixed, 'spacer-1'), [mixed[0]])
    const updated = updateComponentById(list, 'table-0', { tag: 'a2' })
    assert.equal(updated[0].tag, 'a2')
    assert.equal(list[0].tag, 'a')
  })

  it('ignores cross-kind ids instead of acting by bare index', () => {
    const list = [{ type: 'spacer', height: 1 }, { type: 'spacer', height: 2 }]
    assert.equal(deleteComponentById(list, 'table-0'), list)
    const updated = updateComponentById(list, 'image-0', { tag: 'x' })
    assert.equal(updated, list)
    assert.ok(!('tag' in updated[0]))
    assert.equal(findElementById({ components: [{ type: 'table', tag: 't' }] }, 'spacer-0'), null)
  })

  it('moves and reorders components with boundary guards', () => {
    const list = [table('a'), table('b'), table('c')]
    assert.deepEqual(moveComponentByIndex(list, 1, 'up').map((c) => c.tag), ['b', 'a', 'c'])
    assert.deepEqual(moveComponentByIndex(list, 1, 'down').map((c) => c.tag), ['a', 'c', 'b'])
    assert.deepEqual(moveComponentByIndex(list, 0, 'up').map((c) => c.tag), ['a', 'b', 'c'])
    assert.deepEqual(reorderComponentsByIndex(list, 0, 2).map((c) => c.tag), ['b', 'c', 'a'])
    assert.equal(reorderComponentsByIndex(list, 1, 1), list)
    assert.equal(reorderComponentsByIndex(list, 0, 9), list)
  })

  it('rejects fractional indices instead of polluting the array', () => {
    const list = ['a', 'b']
    const moved = moveComponentByIndex(list, 0.5, 'up')
    assert.deepEqual(moved, ['a', 'b'])
    assert.ok(!('0.5' in moved) && !('-0.5' in moved))
  })

  it('pastes after the target id and clones do not share identity', () => {
    const clone = { type: 'spacer', height: 7 }
    const once = pasteComponentAt([table('a')], clone, 'table-0')
    assert.equal(once.length, 2)
    assert.deepEqual(once[1], clone)
    assert.notEqual(once[1], clone)
    const twice = pasteComponentAt(pasteComponentAt([], clone, null), clone, null)
    twice[0].height = 99
    assert.equal(twice[1].height, 7)
  })
})
