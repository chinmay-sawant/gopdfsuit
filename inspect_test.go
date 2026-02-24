package form

import (
"fmt"
"io/ioutil"
"regexp"
"testing"
)

func TestWidgetHierarchy(t *testing.T) {
	b, _ := ioutil.ReadFile("../../../sampledata/filler/jefferson/jefferson.pdf")
	
	// Print object 615 and 253 to see what they are (kids of MRN)
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	for _, m := range objRe.FindAllSubmatch(b, -1) {
		objNum := string(m[1])
		if objNum == "615" || objNum == "253" {
			fmt.Printf("Object %s:\n%s\n---\n", objNum, string(m[3]))
		}
	}
}
