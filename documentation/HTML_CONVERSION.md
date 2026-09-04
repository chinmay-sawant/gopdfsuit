# HTML Conversion - Pure-Go Engine (gowkhtmltopdf)

`POST /api/v1/htmltopdf` and `POST /api/v1/htmltoimage` run on
`github.com/chinmay-sawant/gowkhtmltopdf` (pure-Go, no browser, no CGO).
There is no Chromium, headless-shell, or `CHROME_PATH` runtime.

## Fidelity limits

- Scripts are stripped: JS-heavy SPAs render as static HTML only. Expect
  missing content on pages that require JS to build DOM.
- Partial flex/grid support: complex modern layouts may reflow or stack.
- Backgrounds and gradients ignored: `background` CSS does not paint.
- Fonts: WOFF2 and data-URI `@font-face` are skipped. Prefer TTF/WOFF or
  system fonts.
- Images: SVG output is not supported. `htmltoimage` accepts png/jpg only.

## No-op parameters (kept for compatibility, ignored by engine)

- `htmltopdf`: `dpi`, `low_quality` are no-ops and hidden in the UI.
- `htmltoimage`: `zoom`, `crop_width`, `crop_height`, `crop_x`, `crop_y`
  are no-ops and hidden in the UI. `format: svg` is rejected.

## Supported knobs

- `htmltopdf`: `page_size`, `orientation`, `margin_*`, `grayscale`.
- `htmltoimage`: `format` (png/jpg), `width`, `height`, `quality`.
