# Typst Math Syntax Reference

> Extracted from the official [Typst Math Documentation](https://typst.app/docs/reference/math/).
> This file serves as the parsing reference when math mode is enabled in table cells.

---

## 1. Equation (Core)

**Doc**: https://typst.app/docs/reference/math/equation/

Math expressions are wrapped in dollar signs `$...$`. Adding whitespace before/after lifts into block mode.

### Syntax

```typst
$ A = pi r^2 $                          // inline math
$ a^2 + b^2 = c^2 $                     // block math (with spaces)
$ sum_(k=1)^n k = (n(n+1)) / 2 $        // summation with limits
```

### Key Rules

- Single letters → displayed as-is
- Multiple letters → interpreted as variables/functions
- Quotes `"text"` → verbatim text
- `#var` → access code-level variables
- `&` → alignment point
- `\\` → line break in equations

### Parameters

| Parameter      | Type                            | Description                     |
| -------------- | ------------------------------- | ------------------------------- |
| `block`        | `bool`                          | Whether equation is block-level |
| `numbering`    | `none\|str\|function`           | Equation numbering              |
| `number-align` | `alignment`                     | Alignment of equation number    |
| `supplement`   | `none\|auto\|content\|function` | Supplement for references       |
| `alt`          | `none\|str`                     | Alt text for accessibility      |
| `body`         | `content`                       | The math content                |

---

## 2. Accent

**Doc**: https://typst.app/docs/reference/math/accent/

Adds diacritical marks above math content.

### Syntax

```typst
$ grave(a) = accent(a, `) $
$ arrow(a) = accent(a, arrow) $
$ tilde(a) = accent(a, \u{0303}) $
```

### Shorthand Functions

`grave()`, `acute()`, `hat()`, `tilde()`, `macron()`, `breve()`, `dot()`, `ddot()`, `dddot()`, `ddddot()`, `circle()`, `arrow()`, `bar()`

### Parameters

| Parameter | Type           | Description                                 |
| --------- | -------------- | ------------------------------------------- |
| `base`    | `content`      | Content to accent (positional, required)    |
| `accent`  | `str\|content` | The accent character (positional, required) |
| `size`    | `relative`     | Size of the accent                          |
| `dotless` | `bool`         | Whether to remove dot from i/j              |

---

## 3. Attach (Subscript / Superscript)

**Doc**: https://typst.app/docs/reference/math/attach/

Adds attachments (subscripts/superscripts) to expressions.

### Syntax

```typst
$ sum_(i=0)^n a_i = 2^(1+i) $
```

### Dedicated Syntax

- `_` → subscript (bottom attachment)
- `^` → superscript (top attachment)

### Parameters

| Parameter | Type            | Description                         |
| --------- | --------------- | ----------------------------------- |
| `base`    | `content`       | Base content (positional, required) |
| `t`       | `none\|content` | Top/superscript                     |
| `b`       | `none\|content` | Bottom/subscript                    |
| `tl`      | `none\|content` | Top-left                            |
| `bl`      | `none\|content` | Bottom-left                         |
| `tr`      | `none\|content` | Top-right                           |
| `br`      | `none\|content` | Bottom-right                        |

### Related Functions

- `scripts()` → force scripts-style attachment
- `limits()` → force limits-style attachment

---

## 4. Binomial

**Doc**: https://typst.app/docs/reference/math/binom/

Binomial coefficient notation.

### Syntax

```typst
$ binom(n, k) $
$ binom(n, k_1, k_2, k_3, ..., k_m) $
```

### Parameters

| Parameter | Type      | Description                                    |
| --------- | --------- | ---------------------------------------------- |
| `upper`   | `content` | Upper part (positional, required)              |
| `lower`   | `content` | Lower part(s) (positional, required, variadic) |

---

## 5. Cancel

**Doc**: https://typst.app/docs/reference/math/cancel/

Draws a cancellation line through content.

### Syntax

```typst
$ (a dot b dot cancel(x)) / cancel(x) $
```

### Parameters

| Parameter  | Type                    | Description                              |
| ---------- | ----------------------- | ---------------------------------------- |
| `body`     | `content`               | Content to cancel (positional, required) |
| `length`   | `relative`              | Length of the cancel line                |
| `inverted` | `bool`                  | Invert the cancel line direction         |
| `cross`    | `bool`                  | Draw an X instead of a single line       |
| `angle`    | `auto\|angle\|function` | Angle of the cancel line                 |
| `stroke`   | `stroke`                | Stroke style for the line                |

---

## 6. Cases

**Doc**: https://typst.app/docs/reference/math/cases/

Piecewise-defined functions with a brace.

### Syntax

```typst
$ f(x, y) := cases(
  1 "if" (x dot y)/2 <= 0,
  2 "if" x "is even",
  3 "if" x in NN,
  4 "else",
) $
```

### Parameters

| Parameter  | Type                       | Description                                   |
| ---------- | -------------------------- | --------------------------------------------- |
| `delim`    | `none\|str\|array\|symbol` | Delimiter character                           |
| `reverse`  | `bool`                     | Reverse the delimiter position                |
| `gap`      | `relative`                 | Gap between cases                             |
| `children` | `content`                  | The branches (positional, required, variadic) |

---

## 7. Class

**Doc**: https://typst.app/docs/reference/math/class/

Assigns a math class to content, affecting spacing.

### Syntax

```typst
#let loves = math.class("relation", sym.suit.heart)
$ x loves y and y loves 5 $
```

### Classes

`"normal"`, `"punctuation"`, `"opening"`, `"closing"`, `"fence"`, `"large"`, `"relation"`, `"unary"`, `"binary"`, `"vary"`

### Parameters

| Parameter | Type      | Description                                |
| --------- | --------- | ------------------------------------------ |
| `class`   | `str`     | The class name (positional, required)      |
| `body`    | `content` | Content to classify (positional, required) |

---

## 8. Fraction

**Doc**: https://typst.app/docs/reference/math/frac/

Fraction notation (numerator over denominator).

### Syntax

```typst
$ 1/2 < (x+1)/2 $
$ ((x+1)) / 2 = frac(a, b) $
```

### Dedicated Syntax

Use `/` between expressions to create a fraction. Use parentheses to group multi-atom expressions.

### Parameters

| Parameter | Type      | Description                             |
| --------- | --------- | --------------------------------------- |
| `num`     | `content` | Numerator (positional, required)        |
| `denom`   | `content` | Denominator (positional, required)      |
| `style`   | `str`     | Fraction style: `"inline"`, `"display"` |

---

## 9. Left/Right (Delimiters)

**Doc**: https://typst.app/docs/reference/math/lr/

Auto-scaling delimiters and bracket functions.

### Syntax

```typst
$ [a, b/2] $
$ lr(]sum_(x=1)^n], size: #50%) x $
$ abs((x + y) / 2) $
$ \{ (x / y) \} $
```

### Built-in Functions

| Function  | Description                        |
| --------- | ---------------------------------- |
| `lr()`    | Auto-scaling left/right delimiters |
| `mid()`   | Mid delimiter (e.g., `\|`)         |
| `abs()`   | Absolute value `\|...\|`           |
| `norm()`  | Norm `‖...‖`                       |
| `floor()` | Floor brackets `⌊...⌋`             |
| `ceil()`  | Ceiling brackets `⌈...⌉`           |
| `round()` | Rounding brackets `⌊...⌉`          |

### Parameters (for `lr`)

| Parameter | Type       | Description                                      |
| --------- | ---------- | ------------------------------------------------ |
| `body`    | `content`  | Content within delimiters (positional, required) |
| `size`    | `relative` | Size of delimiters                               |

---

## 10. Matrix

**Doc**: https://typst.app/docs/reference/math/mat/

Matrix notation with rows and columns.

### Syntax

```typst
$ mat(
  1, 2, ..., 10;
  2, 2, ..., 10;
  dots.v, dots.v, dots.down, dots.v;
  10, 10, ..., 10;
) $
```

### Key Rules

- Commas `,` separate columns
- Semicolons `;` separate rows

### Parameters

| Parameter    | Type                       | Description                               |
| ------------ | -------------------------- | ----------------------------------------- |
| `delim`      | `none\|str\|array\|symbol` | Delimiter around matrix                   |
| `align`      | `alignment`                | Content alignment                         |
| `augment`    | `none\|int\|dictionary`    | Augmented matrix divider                  |
| `gap`        | `relative`                 | Gap between cells                         |
| `row-gap`    | `relative`                 | Gap between rows                          |
| `column-gap` | `relative`                 | Gap between columns                       |
| `rows`       | `array`                    | Row data (positional, required, variadic) |

---

## 11. Primes

**Doc**: https://typst.app/docs/reference/math/primes/

Prime notation using apostrophes.

### Syntax

```typst
$ a' $        // single prime
$ a'' $       // double prime
$ a''' $      // triple prime
```

### Dedicated Syntax

Apostrophes `'` are used as primes. They attach to the previous element automatically.

### Parameters

| Parameter | Type  | Description                             |
| --------- | ----- | --------------------------------------- |
| `count`   | `int` | Number of primes (positional, required) |

---

## 12. Roots

**Doc**: https://typst.app/docs/reference/math/roots/

Square roots and nth roots.

### Syntax

```typst
$ sqrt(3 - 2 sqrt(2)) = sqrt(2) - 1 $
$ root(3, x) $
```

### Functions

| Function               | Description |
| ---------------------- | ----------- |
| `sqrt(content)`        | Square root |
| `root(index, content)` | Nth root    |

### Parameters (for `root`)

| Parameter  | Type            | Description                                      |
| ---------- | --------------- | ------------------------------------------------ |
| `index`    | `none\|content` | Root index                                       |
| `radicand` | `content`       | Expression under the root (positional, required) |

---

## 13. Sizes

**Doc**: https://typst.app/docs/reference/math/sizes/

Force display/inline/script size styling.

### Syntax

```typst
$ sum_i x_i/2 = display(sum_i x_i/2) $
$ sum_i x_i/2 = inline(sum_i x_i/2) $
```

### Functions

| Function        | Description                                 |
| --------------- | ------------------------------------------- |
| `display(body)` | Force display size (normal block)           |
| `inline(body)`  | Force inline size (normal text)             |
| `script(body)`  | Force script size (subscript level)         |
| `sscript(body)` | Force sub-script size (sub-subscript level) |

### Parameters

| Parameter | Type      | Description                                 |
| --------- | --------- | ------------------------------------------- |
| `body`    | `content` | Content to size (positional, required)      |
| `cramped` | `bool`    | Restrict exponent height (default: `false`) |

---

## 14. Stretch

**Doc**: https://typst.app/docs/reference/math/stretch/

Stretches content horizontally or vertically.

### Syntax

```typst
$ stretch(=, size: #2em) $
```

### Parameters

| Parameter | Type       | Description                               |
| --------- | ---------- | ----------------------------------------- |
| `body`    | `content`  | Content to stretch (positional, required) |
| `size`    | `relative` | Target stretch size                       |

---

## 15. Styles (Bold / Italic / Upright)

**Doc**: https://typst.app/docs/reference/math/styles/

Font style functions for math content.

### Syntax

```typst
$ upright(A) != A $
$ bold(A) != A $
$ italic(x) $
```

### Functions

| Function        | Description                                      |
| --------------- | ------------------------------------------------ |
| `upright(body)` | Non-italic (roman) style                         |
| `italic(body)`  | Italic style (default for roman/greek lowercase) |
| `bold(body)`    | Bold style                                       |

---

## 16. Text Operator

**Doc**: https://typst.app/docs/reference/math/op/

Upright text operators in math (like sin, cos, etc.).

### Syntax

```typst
$ tan x = (sin x)/(cos x) $
$ op("custom", limits: #true)_(n->oo) n $
```

### Predefined Operators

`arccos`, `arcsin`, `arctan`, `arg`, `cos`, `cosh`, `cot`, `coth`, `csc`, `csch`, `ctg`, `deg`, `det`, `dim`, `exp`, `gcd`, `lcm`, `hom`, `id`, `im`, `inf`, `ker`, `lg`, `lim`, `liminf`, `limsup`, `ln`, `log`, `max`, `min`, `mod`, `Pr`, `sec`, `sech`, `sin`, `sinc`, `sinh`, `sup`, `tan`, `tanh`, `tg`, `tr`

### Parameters

| Parameter | Type      | Description                          |
| --------- | --------- | ------------------------------------ |
| `text`    | `content` | Operator text (positional, required) |
| `limits`  | `bool`    | Whether to show limits above/below   |

---

## 17. Under/Over

**Doc**: https://typst.app/docs/reference/math/underover/

Lines and braces above/below content.

### Syntax

```typst
$ underline(1 + 2 + ... + 5) $
$ overline(1 + 2 + ... + 5) $
$ underbrace(0 + 1 + dots.c + n, n + 1 "numbers") $
$ overbrace(0 + 1 + dots.c + n, n + 1 "numbers") $
```

### Functions

| Function                          | Description                          |
| --------------------------------- | ------------------------------------ |
| `underline(body)`                 | Line under content                   |
| `overline(body)`                  | Line over content                    |
| `underbrace(body, annotation?)`   | Brace under with optional annotation |
| `overbrace(body, annotation?)`    | Brace over with optional annotation  |
| `underbracket(body, annotation?)` | Bracket under                        |
| `overbracket(body, annotation?)`  | Bracket over                         |
| `underparen(body, annotation?)`   | Parenthesis under                    |
| `overparen(body, annotation?)`    | Parenthesis over                     |
| `undershell(body, annotation?)`   | Shell under                          |
| `overshell(body, annotation?)`    | Shell over                           |

---

## 18. Variants (Font Variants)

**Doc**: https://typst.app/docs/reference/math/variants/

Different mathematical font styles.

### Syntax

```typst
$ sans(A B C) $
$ frak(P) $
$ mono(x + y = z) $
$ bb(b) $
$ bb(N) = NN $
$ f: NN -> RR $
$ cal(A) $
```

### Functions

| Function      | Description                     |
| ------------- | ------------------------------- |
| `serif(body)` | Serif/roman font (default)      |
| `sans(body)`  | Sans-serif font                 |
| `frak(body)`  | Fraktur font                    |
| `mono(body)`  | Monospace font                  |
| `bb(body)`    | Blackboard bold (double-struck) |
| `cal(body)`   | Calligraphic font               |
| `scr(body)`   | Script font                     |

### Special Symbols

Blackboard bold uppercase: `NN` (naturals), `ZZ` (integers), `QQ` (rationals), `RR` (reals), `CC` (complex)

---

## 19. Vector

**Doc**: https://typst.app/docs/reference/math/vec/

Column vector notation.

### Syntax

```typst
$ vec(a, b, c) dot vec(1, 2, 3) = a + 2b + 3c $
```

### Parameters

| Parameter  | Type                       | Description                                      |
| ---------- | -------------------------- | ------------------------------------------------ |
| `delim`    | `none\|str\|array\|symbol` | Delimiter character                              |
| `align`    | `alignment`                | Content alignment                                |
| `gap`      | `relative`                 | Gap between elements                             |
| `children` | `content`                  | Vector elements (positional, required, variadic) |

---

## Common Symbols Reference

| Symbol            | Typst Syntax | Rendered |
| ----------------- | ------------ | -------- |
| Greek pi          | `pi`         | π        |
| Greek alpha       | `alpha`      | α        |
| Greek beta        | `beta`       | β        |
| Summation         | `sum`        | Σ        |
| Product           | `product`    | Π        |
| Integral          | `integral`   | ∫        |
| Infinity          | `oo`         | ∞        |
| Natural numbers   | `NN`         | ℕ        |
| Real numbers      | `RR`         | ℝ        |
| Integer numbers   | `ZZ`         | ℤ        |
| Rational numbers  | `QQ`         | ℚ        |
| Complex numbers   | `CC`         | ℂ        |
| Dot product       | `dot`        | ·        |
| Cross product     | `times`      | ×        |
| Not equal         | `!=`         | ≠        |
| Less/equal        | `<=`         | ≤        |
| Greater/equal     | `>=`         | ≥        |
| Implies           | `=>`         | ⇒        |
| Element of        | `in`         | ∈        |
| Not element of    | `in.not`     | ∉        |
| Subset            | `subset`     | ⊂        |
| Superset          | `supset`     | ⊃        |
| Dots (horizontal) | `dots.c`     | ⋯        |
| Dots (vertical)   | `dots.v`     | ⋮        |
| Dots (diagonal)   | `dots.down`  | ⋱        |
| Arrow right       | `->`         | →        |
| Arrow left        | `<-`         | ←        |
| For all           | `forall`     | ∀        |
| Exists            | `exists`     | ∃        |
