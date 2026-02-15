# Typst Math Syntax Module

## Purpose

This module provides Typst math syntax support for rendering mathematical expressions in PDF table cells. When the **math mode** option is enabled in the config (frontend + model), table cell values containing Typst math syntax (e.g., `$ A = pi r^2 $`) will be parsed and rendered as proper mathematical notation.

## How It Works

1. **Detection**: Cell values wrapped in `$...$` are detected as math expressions
2. **Parsing**: The Typst math syntax is parsed into an AST (Abstract Syntax Tree)
3. **Rendering**: The AST is converted into PDF-compatible math content

## Input Format

When a table cell contains math-enabled content, the value follows Typst syntax:

```
$ A = pi r^2 $           → renders: A = πr²
$ frac(a, b) $            → renders: a/b (fraction)
$ sqrt(x) $               → renders: √x
$ sum_(i=1)^n i $         → renders: Σᵢ₌₁ⁿ i
$ vec(a, b, c) $          → renders: column vector (a, b, c)
$ mat(1, 2; 3, 4) $       → renders: 2×2 matrix
```

## Config Toggle

The math rendering feature is controlled by a config flag:

```json
{
  "mathEnabled": true
}
```

When `mathEnabled` is `false`, the raw text is rendered as-is without parsing.

## Reference Files

- [SYNTAX_REFERENCE.md](./SYNTAX_REFERENCE.md) — Complete Typst math syntax reference with all functions, parameters, and examples

## Source Documentation

All syntax was extracted from the official Typst documentation:

- https://typst.app/docs/reference/math/
