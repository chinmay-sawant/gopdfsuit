// Cross-tier golden test (E2): the canonical sample template must parse/build
// identically through the Editor document model AND satisfy
// frontend/template.schema.json. Run: npm run test:schema
import { readFile } from 'node:fs/promises'

import {
  buildTemplate,
  buildTemplateJson,
  parseTemplateData,
  validateTemplate,
  PROPS_PATTERN,
  MARGIN_PATTERN,
} from '../src/components/editor/documentModel.js'

const root = new URL('../../', import.meta.url)
const readJson = async (rel) => JSON.parse(await readFile(new URL(rel, root), 'utf8'))

const assert = (cond, message) => {
  if (!cond) {
    console.error(`FAIL: ${message}`)
    process.exitCode = 1
  } else {
    console.log(`ok: ${message}`)
  }
}

const schema = await readJson('frontend/template.schema.json')
const sample = await readJson('sampledata/editor/financial_report.json')

assert(schema.title === 'GoPdfSuit PDFTemplate', 'schema targets PDFTemplate')
assert(schema.definitions?.props?.pattern === PROPS_PATTERN.source, 'schema props pattern matches document model')
assert(schema.definitions?.config === undefined, 'no stray config definition')

const schemaProps = new RegExp(schema.definitions.props.pattern)
const collectProps = (node, out) => {
  if (Array.isArray(node)) {
    node.forEach((entry) => collectProps(entry, out))
    return
  }
  if (node && typeof node === 'object') {
    for (const [key, value] of Object.entries(node)) {
      if (key === 'props' && typeof value === 'string') out.push(value)
      else collectProps(value, out)
    }
  }
}
const sampleProps = []
collectProps(sample, sampleProps)
assert(sampleProps.length > 0, `sample has props strings (${sampleProps.length})`)
assert(sampleProps.every((p) => schemaProps.test(p)), 'every sample props string matches schema grammar')

const first = parseTemplateData(sample)
assert(first.components.length === sample.elements.length, 'parse preserves element count')
const builtOnce = buildTemplate(first)
const checkOnce = validateTemplate(builtOnce)
assert(checkOnce.errors.length === 0, `built template has no schema errors (${checkOnce.errors.join('; ')})`)
assert(checkOnce.warnings.length === 0, `built template has no schema warnings (${checkOnce.warnings.join('; ')})`)

const rebuilt = buildTemplate(parseTemplateData(JSON.parse(buildTemplateJson(first))))
assert(JSON.stringify(rebuilt) === JSON.stringify(builtOnce), 'parse/build round-trip is stable')

if (first.config.pageMargin) {
  assert(MARGIN_PATTERN.test(first.config.pageMargin), 'pageMargin matches margin grammar')
}
const tables = first.components.filter((c) => c.type === 'table')
assert(tables.length > 0 && tables.every((t) => Number.isInteger(t.maxcolumns) && t.maxcolumns > 0), 'tables carry positive integer maxcolumns')

if (process.exitCode) console.error('schema self-check FAILED')
else console.log('schema self-check PASSED')
