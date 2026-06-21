//go:build ignore
package main
import "fmt"
func main() {
	t := buildActiveTraderTemplate()
	for i, e := range t.Elements {
		if e.Table == nil || len(e.Table.Rows) < 3 { continue }
		tbl := e.Table
		templateRow := 1
		template := tbl.Rows[templateRow]
		ok := true
		for ri := templateRow + 1; ri < len(tbl.Rows); ri++ {
			row := tbl.Rows[ri]
			ncol := min(len(row.Row), len(template.Row), tbl.MaxColumns)
			for ci := range ncol {
				tc := template.Row[ci]
				c := row.Row[ci]
				if tc.Props != c.Props {
					ok = false
					fmt.Printf("elem %d row %d col %d props mismatch\n", i, ri, ci)
				}
			}
		}
		fmt.Printf("elem %d rows=%d cols=%d SharedRowLayout_eligible=%v\n", i, len(tbl.Rows), tbl.MaxColumns, ok)
	}
}
