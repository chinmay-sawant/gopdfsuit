package signature

import (
	"bytes"
	"testing"
)

func TestBuildSignaturePlacementMatchesLocate(t *testing.T) {
	var sigValue bytes.Buffer
	sigValue.WriteString("<< /Type /Sig")
	sigValue.WriteByte(' ')
	ids := &SignatureIDs{}
	ids.ByteRangeRel = sigValue.Len()
	sigValue.WriteString(sigByteRangePlaceholder)
	sigValue.WriteString(" /Contents <")
	ids.ContentsDataStartRel = sigValue.Len()
	for range sigContentsHexLen {
		sigValue.WriteByte('0')
	}
	ids.ContentsDataEndRel = sigValue.Len()
	sigValue.WriteString(">>")

	pdf := append([]byte("%PDF-2.0\n"), sigValue.Bytes()...)
	objBodyStart := len("%PDF-2.0\n")

	built := buildSignaturePlacement(objBodyStart, ids, len(pdf))
	located, err := locateSignaturePlacement(pdf)
	if err != nil {
		t.Fatal(err)
	}

	if built.byteRangePos != located.byteRangePos {
		t.Fatalf("byteRangePos built=%d located=%d", built.byteRangePos, located.byteRangePos)
	}
	if built.contentsStart != located.contentsStart || built.contentsEnd != located.contentsEnd {
		t.Fatalf("contents range mismatch built=%d..%d located=%d..%d",
			built.contentsStart, built.contentsEnd, located.contentsStart, located.contentsEnd)
	}
	if built.byteRange != located.byteRange {
		t.Fatalf("byteRange mismatch built=%v located=%v", built.byteRange, located.byteRange)
	}
}