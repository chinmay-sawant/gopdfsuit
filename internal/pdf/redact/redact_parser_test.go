package redact

import (
	"bytes"
	"testing"
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
