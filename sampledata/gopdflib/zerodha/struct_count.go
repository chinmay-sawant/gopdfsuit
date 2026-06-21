//go:build ignore
package main
import (
	"fmt"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)
func estStructCount(t gopdflib.PDFTemplate) int {
	count := 1
	if t.Title.Text != "" { count++ }
	if t.Title.Table != nil {
		count += 1 + len(t.Title.Table.Rows)
		for _, row := range t.Title.Table.Rows { count += len(row.Row) }
	}
	for _, table := range t.Table {
		count += 1 + len(table.Rows)
		for _, row := range table.Rows { count += min(len(row.Row), table.MaxColumns) }
	}
	for _, elem := range t.Elements {
		if elem.Table == nil { continue }
		count += 1 + len(elem.Table.Rows)
		for _, row := range elem.Table.Rows { count += min(len(row.Row), elem.Table.MaxColumns) }
	}
	if t.Footer.Text != "" { count++ }
	return count
}
func main() {
	for name, t := range map[string]gopdflib.PDFTemplate{
		"retail": buildRetailTemplate(),
		"active": buildActiveTraderTemplate(),
		"hft": buildHFTTemplate(),
	} {
		fmt.Printf("%s: struct_elems≈%d arena_eligible=%v signing=%v\n", name, estStructCount(t), estStructCount(t)>=512, t.Config.Signature!=nil && t.Config.Signature.Enabled)
	}
}
