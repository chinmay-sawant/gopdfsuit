# Plans/gopdflib - Fluent Font plus Cell Builder

> **Parent:** `skills/phase-wise-checklist/SKILL.md` - canonical ledger shape; refs `pkg/gopdflib/builder.go`, `pkg/gopdflib/props.go`, `bindings/python/pypdfsuit/builder.py`
> **Status:** implemented - fluent chains shipped in Go plus Python, docs and snippets rewritten, old API byte-identical
> **Estimated effort:** S - one new Go file plus one Python module plus docs, old API untouched

---

## Overview

Add an idiomatic fluent builder over the existing Props grammar so callers write `Font("Helvetica").Size(12).Bold().Center().Props()` (Go) and `Font("Helvetica").size(12).bold().center().props()` (Python) instead of hand-assembling colon-separated strings. The grammar, engine, and wire format do not change; the builder emits the same strings `FontOpts.String` produces.

## Executive Summary

`pkg/gopdflib/builder.go:169-186` already has `FontOpts` plus `MakeProps`, but every call site still passes either a raw string or a 7-arg `MakeProps` call (`builder-snippets/main.go:28-44`, `doc.go:72-82`). `bindings/python/pypdfsuit/builder.py:45-80` mirrors the same stringly API (`make_props`, `new_cell` default `"Helvetica:12:000:left:1:1:1:1"`). The plan adds a chainable layer on top in both languages with byte-identical output, proven by round-trip tests against `MakeProps`/`make_props`.

## Phase 1: Go fluent builder (no behavior change)

- [x] `pkg/gopdflib/fontbuilder.go` (new) - `Font(name)` with `Size/Bold/Italic/Underline/Left/Center/Right/Borders/Bordered/Borderless`, terminal `Props()` plus `Cell(text)`; `Text(s)` with `WithFont/Props/Bg/Fg/Math`, terminal `Build()` - proof: `TestFontBuilderRoundTrip`, `TestFontBuilderCellTerminal`, `TestFluentCellBuilder`, `TestFontBuilderNilSafety` green
- [x] Round-trip plus nil-safety tests - proof: same run green; zero-value default documented (`Font("Helvetica").Props()` equals `MakeProps("Helvetica", 0, ...)`)

## Phase 2: Go docs plus snippets

- [x] `pkg/gopdflib/doc.go:72-82` - lead examples use the fluent chain, raw strings kept as low-level form - proof: full `go test ./pkg/gopdflib` green
- [x] `sampledata/gopdflib/builder-snippets/main.go:28-44` - same rewrite - proof: snippet builds
- [x] `documentation/` template reference - fluent section in `TEMPLATE_REFERENCE.md` - proof: doc diff

## Phase 3: pypdfsuit mirror (snake_case, same chains)

- [x] `bindings/python/pypdfsuit/builder.py` - `Font` plus `Text` chains with `props()`/`cell()`/`build()` terminals - proof: `test_builder_fluent.py` 11 passed
- [x] `new_cell` gains `font=` (explicit `props` wins, documented) - proof: precedence test green
- [x] `__init__.py` exports `Font`, `Text` - proof: import check green
- [x] Parity tests - proof: 11 passed

## Phase 4: Closure gates

- [x] `go build ./...` plus `go test ./pkg/gopdflib -count=1` plus `pytest bindings/python -k builder` - proof: BUILD_OK, pkg ok 0.642s, 11 passed
- [x] Old API untouched: `NewCell(text, props)`, `MakeProps`, `make_props`, `new_cell(text, props)` signatures and outputs byte-identical - proof: pre-existing tests unmodified and green

## Dependencies

- Grammar source of truth: `pkg/gopdflib/props.go:45-53` (`FontOpts`, `styleCode`, `String`, `ParseFontOpts`)
- Go surface: `pkg/gopdflib/builder.go:169-324`, `doc.go:72-82`, `sampledata/gopdflib/builder-snippets/main.go`
- Python surface: `bindings/python/pypdfsuit/{builder.py:45-80,__init__.py:13-127}`
- Explicit non-goals: grammar changes, engine draw changes, `SetCell*` rewrites, TypeScript/frontend API
