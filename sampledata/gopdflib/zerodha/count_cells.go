//go:build ignore
package main
import (
	"fmt"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)
func countCells(t gopdflib.PDFTemplate) (cells, maxRows, tables int) {
	if t.Title.Table != nil {
		tables++
		for _, r := range t.Title.Table.Rows {
			cells += len(r.Row)
		}
		maxRows = max(maxRows, len(t.Title.Table.Rows))
	}
	for _, e := range t.Elements {
		if e.Table != nil {
			tables++
			for _, r := range e.Table.Rows {
				cells += len(r.Row)
			}
			if len(e.Table.Rows) > maxRows {
				maxRows = len(e.Table.Rows)
			}
		}
	}
	return
}
func main() {
	benchCompliant = true
	for name, t := range map[string]gopdflib.PDFTemplate{
		"retail": buildRetailTemplate(),
		"active": buildActiveTraderTemplate(),
		"hft": buildHFTTemplate(),
	} {
		c, mr, tb := countCells(t)
		fmt.Printf("%s: tables=%d total_cells=%d max_table_rows=%d shared=%v\n", name, tb, c, mr, hasShared(t))
	}
}
func hasShared(t gopdflib.PDFTemplate) bool {
	for _, e := range t.Elements {
		if e.Table != nil && e.Table.SharedRowLayout { return true }
	}
	return false
}
