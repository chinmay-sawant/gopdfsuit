package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: stripap <pdf-file>")
		os.Exit(1)
	}
	inPath := os.Args[1]
	b, err := ioutil.ReadFile(inPath)
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	// truncate at last startxref to remove old xref/trailer
	sx := bytes.LastIndex(b, []byte("startxref"))
	base := b
	if sx >= 0 {
		base = b[:sx]
	}

	out := make([]byte, len(base))
	copy(out, base)

	// Remove /AP << ... >> (non-greedy, multi-line) and also /AP <num> 0 R
	apDictRe := regexp.MustCompile(`(?s)/AP\s*<<.*?>>`)
	apRefRe := regexp.MustCompile(`/AP\s*\d+\s+0\s+R`)

	out = apDictRe.ReplaceAll(out, []byte(" "))
	out = apRefRe.ReplaceAll(out, []byte(" "))

	// Rebuild xref from objects present in new bytes
	objRe := regexp.MustCompile(`(\d+)\s+0\s+obj`)
	allObjMatches := objRe.FindAllSubmatchIndex(out, -1)
	offsets := make(map[int]int)
	maxObj := 0
	for _, m := range allObjMatches {
		numBytes := out[m[2]:m[3]]
		if n, err := strconv.Atoi(string(numBytes)); err == nil {
			offsets[n] = m[0]
			if n > maxObj {
				maxObj = n
			}
		}
	}

	xrefStart := len(out)
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", maxObj+1))
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
		} else {
			buf.WriteString("0000000000 00000 f \n")
		}
	}

	// Attempt to detect Root object number from original bytes
	root := 1
	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s0\sR`)
	if rm := rootRe.FindSubmatch(b); len(rm) > 1 {
		if r, err := strconv.Atoi(string(rm[1])); err == nil {
			root = r
		}
	}

	trailer := fmt.Sprintf("trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%EOF\n", maxObj+1, root, xrefStart)

	out = append(out, buf.Bytes()...)
	out = append(out, []byte(trailer)...)

	outPath := inPath[:len(inPath)-4] + "_noAP.pdf"
	if err := ioutil.WriteFile(outPath, out, 0644); err != nil {
		// try alternate name
		outPath = inPath[:len(inPath)-4] + "_noAP_2.pdf"
		if err2 := ioutil.WriteFile(outPath, out, 0644); err2 != nil {
			fmt.Println("write error:", err2)
			os.Exit(1)
		}
	}

	fmt.Println("wrote:", outPath)
}
