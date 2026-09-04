package compress

import (
	"bytes"
	"regexp"
	"strconv"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/pdfobj"
)

var namedRefRe = regexp.MustCompile(`/([A-Za-z]+)\s+(\d+)\s+\d+\s+R`)

func stripDocumentMetadata(objects map[int]pdfObject, catalogNum, infoNum int) {
	if infoNum > 0 {
		delete(objects, infoNum)
	}

	if cat, ok := objects[catalogNum]; ok {
		if metaNum, found := dictObjectRef(cat.body, "Metadata"); found {
			delete(objects, metaNum)
		}
		body := cat.body
		body = removeNamedValue(body, "Metadata")
		body = removeNamedValue(body, "PieceInfo")
		body = removeNamedValue(body, "SpiderInfo")
		cat.body = body
		objects[catalogNum] = cat
	}

	for num, obj := range objects {
		if num == catalogNum {
			continue
		}
		if pdfobj.HasSubstring(obj.body, []byte("/Type /Metadata")) ||
			pdfobj.HasSubstring(obj.body, []byte("/Type/Metadata")) {
			delete(objects, num)
			continue
		}
		body := obj.body
		if thumb, found := dictObjectRef(body, "Thumb"); found {
			delete(objects, thumb)
			body = removeNamedValue(body, "Thumb")
		}
		if pdfobj.HasSubstring(body, []byte("/PieceInfo")) {
			body = removeNamedValue(body, "PieceInfo")
		}
		if !bytes.Equal(body, obj.body) {
			obj.body = body
			objects[num] = obj
		}
	}
}

func dictObjectRef(body []byte, name string) (int, bool) {
	for _, m := range namedRefRe.FindAllSubmatch(body, -1) {
		if string(m[1]) == name {
			n, err := strconv.Atoi(string(m[2]))
			if err == nil && n > 0 {
				return n, true
			}
		}
	}
	return 0, false
}
