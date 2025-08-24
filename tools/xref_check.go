// d:\Chinmay_Personal_Projects\GoPdfSuit\tools\xref_check.go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	objStartRx  = regexp.MustCompile(`\b(\d+)\s+(\d+)\s+obj\b`)
	typeXRefRx  = regexp.MustCompile(`/Type\s*/XRef\b`)
	lengthRx    = regexp.MustCompile(`/Length\s+(\d+)\b|/Length\s+(\d+)\s+0\s+R`)
	numberObjRx = regexp.MustCompile(`\b(\d+)\s+0\s+obj\s+(\d+)\s+endobj`)
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: xref_check file.pdf")
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	// Build simple map of direct number objects (id -> value) for resolving /Length n 0 R
	numberObjs := map[string]int{}
	for _, m := range numberObjRx.FindAllSubmatch(data, -1) {
		id := string(m[1])
		val, _ := strconv.Atoi(string(m[2]))
		numberObjs[id] = val
	}

	r := bufio.NewReader(bytes.NewReader(data))
	var (
		lineBuf []byte
	)
	for {
		b, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}
		lineBuf = append(lineBuf, b...)
		// Detect end of object quickly
		if bytes.Contains(lineBuf, []byte("endobj")) || err == io.EOF {
			if objStartRx.Match(lineBuf) && typeXRefRx.Match(lineBuf) {
				objID := objStartRx.FindSubmatch(lineBuf)[1]
				lengthMatch := lengthRx.FindSubmatch(lineBuf)
				if lengthMatch == nil {
					fmt.Printf("XRef object %s missing /Length\n", objID)
				} else {
					if len(lengthMatch[1]) > 0 {
						fmt.Printf("XRef object %s /Length %s (direct)\n", objID, lengthMatch[1])
					} else {
						refID := string(lengthMatch[2])
						val, ok := numberObjs[refID]
						if !ok {
							fmt.Printf("XRef object %s /Length %s 0 R unresolved\n", objID, refID)
						} else if val <= 0 {
							fmt.Printf("XRef object %s /Length %d invalid\n", objID, val)
						} else {
							fmt.Printf("XRef object %s /Length %d (resolved)\n", objID, val)
						}
					}
				}
			}
			lineBuf = lineBuf[:0]
		}
		if err == io.EOF {
			break
		}
	}

	// Basic sanity: count of 'stream' vs 'endstream'
	streams := strings.Count(string(data), "stream")
	endstreams := strings.Count(string(data), "endstream")
	if streams != endstreams {
		fmt.Printf("Warning: stream/endstream mismatch (%d/%d)\n", streams, endstreams)
	}
}
