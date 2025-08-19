package pdf

// The original `pdf.go` was large. It has been split into smaller files by responsibility:
// - types.go        (page size and dimensions)
// - utils.go        (parsing helpers and string escaping)
// - pagemanager.go  (PageManager and page lifecycle)
// - draw.go         (drawing helpers: title, table, footer, watermark)
// - generator.go    (GenerateTemplatePDF and orchestration)
// - xfdf.go         (XFDF parsing and PDF form filling)

// This file intentionally left minimal to keep package build roots simple.
