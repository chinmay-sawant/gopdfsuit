package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

// hftTableShell captures table metadata while deferring row array parsing.
type hftTableShell struct {
	MaxColumns           int             `json:"maxcolumns"`
	ColumnWidths         []float64       `json:"columnwidths,omitempty"`
	RowHeights           []float64       `json:"rowheights,omitempty"`
	BgColor              string          `json:"bgcolor,omitempty"`
	TextColor            string          `json:"textcolor,omitempty"`
	SharedRowLayout      bool            `json:"sharedRowLayout,omitempty"`
	SharedRowTemplateRow int             `json:"sharedRowTemplateRow,omitempty"`
	RowsRaw              json.RawMessage `json:"rows"`
}

type hftElementShell struct {
	Type   string         `json:"type"`
	Index  int            `json:"index,omitempty"`
	Table  *hftTableShell `json:"table,omitempty"`
	Spacer *models.Spacer `json:"spacer,omitempty"`
	Image  *models.Image  `json:"image,omitempty"`
}

type hftPayloadShell struct {
	Config    models.Config     `json:"config"`
	Title     models.Title      `json:"title"`
	Table     []models.Table    `json:"table"`
	Spacer    []models.Spacer   `json:"spacer,omitempty"`
	Image     []models.Image    `json:"image,omitempty"`
	Elements  []hftElementShell `json:"elements,omitempty"`
	Footer    models.Footer     `json:"footer"`
	Bookmarks []models.Bookmark `json:"bookmarks,omitempty"`
}

func decodeHFTPayload(buf []byte, template *models.PDFTemplate) error {
	var shell hftPayloadShell
	if err := sonic.Unmarshal(buf, &shell); err != nil {
		return err
	}
	if err := applyHFTPayloadShell(&shell, template); err != nil {
		return err
	}
	return nil
}

func applyHFTPayloadShell(shell *hftPayloadShell, t *models.PDFTemplate) error {
	t.Config = shell.Config
	t.Title = shell.Title
	t.Table = shell.Table
	t.Spacer = shell.Spacer
	t.Image = shell.Image
	t.Footer = shell.Footer
	t.Bookmarks = shell.Bookmarks

	if len(shell.Elements) == 0 {
		t.Elements = nil
		return nil
	}

	if len(t.Elements) < len(shell.Elements) {
		t.Elements = append(t.Elements, make([]models.Element, len(shell.Elements)-len(t.Elements))...)
	}
	t.Elements = t.Elements[:len(shell.Elements)]

	for i, se := range shell.Elements {
		elem := &t.Elements[i]
		elem.Type = se.Type
		elem.Index = se.Index
		elem.Spacer = se.Spacer
		elem.Image = se.Image

		if se.Table == nil {
			elem.Table = nil
			continue
		}

		var tbl *models.Table
		if elem.Table != nil {
			tbl = elem.Table
		} else {
			tbl = &models.Table{}
			elem.Table = tbl
		}
		tbl.MaxColumns = se.Table.MaxColumns
		tbl.ColumnWidths = se.Table.ColumnWidths
		tbl.RowHeights = se.Table.RowHeights
		tbl.BgColor = se.Table.BgColor
		tbl.TextColor = se.Table.TextColor
		tbl.SharedRowLayout = se.Table.SharedRowLayout
		tbl.SharedRowTemplateRow = se.Table.SharedRowTemplateRow

		if len(se.Table.RowsRaw) == 0 {
			tbl.Rows = tbl.Rows[:0]
			continue
		}
		if err := fillTableRowsFast(se.Table.RowsRaw, tbl); err != nil {
			return fmt.Errorf("elements[%d].table.rows: %w", i, err)
		}
	}
	return nil
}

func fillTableRowsFast(raw json.RawMessage, tbl *models.Table) error {
	if len(raw) == 0 {
		tbl.Rows = tbl.Rows[:0]
		return nil
	}
	// Second pass: JIT-unmarshal rows into pre-sized row/cell slices (Phase 10 prealloc).
	return sonic.Unmarshal(raw, &tbl.Rows)
}