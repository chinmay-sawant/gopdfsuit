package redact

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

func TestBuildObjectMapUsesMergeScanner(t *testing.T) {
	objMap, objGen, err := buildObjectMap(minimalPDF)
	if err != nil {
		t.Fatalf("buildObjectMap failed: %v", err)
	}

	for _, num := range []int{1, 2, 3, 4} {
		body, ok := objMap[num]
		if !ok {
			t.Fatalf("expected object %d in map", num)
		}
		if len(body) == 0 {
			t.Fatalf("expected non-empty body for object %d", num)
		}
		if objGenNum(objGen, num) != 0 {
			t.Fatalf("expected generation 0 for object %d, got %d", num, objGenNum(objGen, num))
		}
	}

	if !bytes.Contains(objMap[3], []byte("/Type /Page")) {
		t.Fatalf("page object body missing /Type /Page")
	}
	if !bytes.Contains(objMap[4], []byte("stream")) {
		t.Fatalf("content stream object missing stream keyword")
	}
}

func TestFindPageObjectByIntKey(t *testing.T) {
	objMap, _, err := buildObjectMap(minimalMultiPagePDF)
	if err != nil {
		t.Fatalf("buildObjectMap failed: %v", err)
	}

	page1, err := findPageObject(objMap, minimalMultiPagePDF, 1)
	if err != nil {
		t.Fatalf("findPageObject page 1 failed: %v", err)
	}
	page2, err := findPageObject(objMap, minimalMultiPagePDF, 2)
	if err != nil {
		t.Fatalf("findPageObject page 2 failed: %v", err)
	}
	if page1 == page2 {
		t.Fatalf("expected distinct page objects, both were %d", page1)
	}
}

func TestFindTextOccurrencesMultiCachesPageExtraction(t *testing.T) {
	r, err := NewRedactor(minimalMultiPagePDF)
	if err != nil {
		t.Fatalf("NewRedactor failed: %v", err)
	}

	if _, err := r.FindTextOccurrencesMulti([]string{"Alpha", "Beta"}); err != nil {
		t.Fatalf("FindTextOccurrencesMulti failed: %v", err)
	}
	if got, want := len(r.pageTextPositions), 2; got != want {
		t.Fatalf("cached page extractions = %d, want %d", got, want)
	}
}

func TestPageWalkRejectsCyclesInSubprocess(t *testing.T) {
	if os.Getenv("GOPDFSUIT_PAGE_CYCLE_HELPER") == "1" {
		objMap := map[int][]byte{
			2: []byte("<< /Type /Pages /Kids [2 0 R] /Count 1 >>"),
		}
		var dims []models.PageDetail
		if err := traversePages(2, objMap, &dims); err == nil || !strings.Contains(err.Error(), "cycle") {
			os.Exit(2)
		}
		if _, err := findPageObject(
			map[int][]byte{
				1: []byte("<< /Type /Catalog /Pages 2 0 R >>"),
				2: []byte("<< /Type /Pages /Kids [2 0 R] /Count 1 >>"),
			},
			[]byte("/Root 1 0 R"),
			1,
		); err == nil || !strings.Contains(err.Error(), "cycle") {
			os.Exit(3)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run", "^TestPageWalkRejectsCyclesInSubprocess$")
	cmd.Env = append(os.Environ(), "GOPDFSUIT_PAGE_CYCLE_HELPER=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("page cycle helper failed: %v\n%s", err, output)
	}
}

func TestPageWalkRejectsExcessiveDepth(t *testing.T) {
	const depth = 1100
	objMap := make(map[int][]byte, depth+1)
	for num := 1; num <= depth; num++ {
		objMap[num] = []byte(fmt.Sprintf("<< /Type /Pages /Kids [%d 0 R] /Count 1 >>", num+1))
	}
	objMap[depth+1] = []byte("<< /Type /Page /MediaBox [0 0 10 10] >>")

	var dims []models.PageDetail
	err := traversePages(1, objMap, &dims)
	if err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatalf("traversePages error = %v, want depth error", err)
	}
}
